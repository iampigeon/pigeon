all: build

.PHONY: build
build: 
	go generate ./...
	go build -i -o bin/scheduler github.com/WiseGrowth/pigeon/cmd/scheduler
