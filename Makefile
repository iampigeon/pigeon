all: build

.PHONY: build
build b: 
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go generate ./...
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -i -o bin/scheduler github.com/WiseGrowth/pigeon/cmd/scheduler
build-osx bosx:
	@CGO_ENABLED=0 go generate ./...
	@CGO_ENABLED=0 go build -i -o bin/scheduler github.com/WiseGrowth/pigeon/cmd/scheduler
