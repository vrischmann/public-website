build:
	templ generate ./...
	go run go.rischmann.fr/website-generator
	rsync -av assets build/.
	rsync -av --include="*.png" --include="*.pdf" --include="*/" --exclude="*" pages/ build/.

watch-build:
	watchexec --print-events -e templ,css,js,md -w pages -w templates -w assets just build

docker_dev: build
	docker compose up --build

deploy: build
	rsync -avz --delete build/. wevo.rischmann.fr:/data/website/.