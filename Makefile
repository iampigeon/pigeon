all: build

.PHONY: build
docker-compose dcb:
	@docker-compose build
docker-compose dc:
	@docker-compose up
docker-compose dck:
	@docker-compose kill -s SIGINT
build b:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go generate ./...
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -i -o bin/scheduler github.com/iampigeon/pigeon/cmd/scheduler
build-osx bx:
	@CGO_ENABLED=0 go generate ./...
	@CGO_ENABLED=0 go build -i -o bin/scheduler github.com/iampigeon/pigeon/cmd/scheduler
run r: build dc
