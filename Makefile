

SERVER = echeclus.uberspace.de
TARGET_DIR = packages/socialrunclubs.de

all:
	@echo "make sync       -> build and upload to socialrunclubs.de"
	@echo "make run-remote -> sync & run remote script"

 .bin/generate-linux: cmd/generate/main.go go.mod
	mkdir -p .bin
	GOOS=linux GOARCH=amd64 go build -o .bin/generate-linux cmd/generate/main.go

.phony: build
build:
	rm -rf .out
	go run cmd/generate/main.go -config local.json

.repo/.git/config:
	git clone https://github.com/flopp/socialrunclubs-de.git .repo

.phony: sync
sync: .repo/.git/config .bin/generate-linux
	(cd .repo && git pull --quiet)
	rsync -a production.json scripts/cronjob.sh .bin/generate-linux $(SERVER):$(TARGET_DIR)/
	rsync -a .repo/ $(SERVER):$(TARGET_DIR)/repo
	ssh $(SERVER) chmod +x $(TARGET_DIR)/cronjob.sh $(TARGET_DIR)/generate-linux

.phony: run-remote
run-remote: sync
	ssh $(SERVER) $(TARGET_DIR)/cronjob.sh
