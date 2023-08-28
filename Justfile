
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
	rsync -av --include="*.png" --include="*.pdf" --include="*/" --exclude="*" pages/ build/.

watch-build:
	watchexec --print-events -e templ,css,js,md -w pages -w templates -w assets just build

docker_dev: build
	docker compose up --build

deploy: build
	rsync -avz --delete build/. wevo.rischmann.fr:/data/website/.
