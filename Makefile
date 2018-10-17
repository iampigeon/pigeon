KEY=pigeon-api-production.pem
SERVER_USER=ec2-user
DOMAIN=api.iampigeon.com

SSH ?= ssh
SCP ?= scp

run r: build dc-kill clean db-build dc

connect_server cs:
	@echo "[ssh] Connecting..."
	@$(SSH) -i $(KEY) $(SERVER_USER)@$(DOMAIN)

copy_makefile cm:
	@echo "[copy] Sending Makefile to server..."
	@$(SCP) -i $(KEY) Makefile $(SERVER_USER)@$(DOMAIN):~
	@echo "[copy] done"

dc_build dcb:
	@echo "[build] Building Docker Compose..."
	@docker-compose build

docker_compose dc:
	@echo "[run] Running Docker Compose..."
	@docker-compose up

dc_kill dck:
	@echo "[kill] Killing Docker Compose..."
	@docker-compose kill -s SIGINT

clean c:
	@echo "[clean] Cleaning Docker Compose..."
	@docker rm -f $(docker ps -a -q)
	@docker rmi -f $(docker images -q)

build b:
	@echo "[build-linux] Building Pigeon..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go generate ./...
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -i -o bin/scheduler github.com/iampigeon/pigeon/cmd/scheduler

build_osx bx:
	@echo "[build-osx] Building Pigeon..."
	@CGO_ENABLED=0 go generate ./...
	@CGO_ENABLED=0 go build -i -o bin/scheduler github.com/iampigeon/pigeon/cmd/scheduler


.PHONY: run connect_server copy_makefile dc_build docker_compose dc_kill clean build build_osx