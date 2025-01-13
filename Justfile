
tool_templ := "github.com/a-h/templ/cmd/templ@latest"

clean:
	rm -rf build

gen-template:
	@printf "\x1b[34m===>\x1b[m  Running templ generate\n"
	@go run {{tool_templ}} generate

build: clean gen-template
	@printf "\x1b[34m===>\x1b[m  Running website-generator generate\n"
	go run go.rischmann.fr/website-generator generate
	rsync -av files build/.

build-dev: clean gen-template
	@printf "\x1b[34m===>\x1b[m  Running website-generator generate --no-assets-versioning\n"
	go run go.rischmann.fr/website-generator generate --no-assets-versioning
	rsync -av files build/.

fmt:
	@printf "\x1b[34m===>\x1b[m  Running go fmt\n"
	go fmt ./...
	@printf "\x1b[34m===>\x1b[m  Running templ fmt\n"
	go run {{tool_templ}} fmt .

convert-images:
	@printf "\x1b[34m===>\x1b[m  Running 'magick convert' for all png files\n"
	fd -e png -x magick {} {.}.avif

watch-convert-images:
	watchexec --print-events -e png -w pages just convert-images

watch-build:
	watchexec --print-events -e templ,css,js,md,avif -w pages -w templates -w assets just build

watch-build-dev:
	watchexec --print-events -e templ,css,js,md,avif -w pages -w templates -w assets just build-dev

docker_dev: build
	docker compose up --build

deploy: build
	rsync -avz --delete build/. wevo.rischmann.fr:/data/website/.
