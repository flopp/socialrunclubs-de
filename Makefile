

SERVER = echeclus.uberspace.de
TARGET_DIR = packages/socialrunclubs.de

all:
	@echo "make check      -> run testing and linting"
	@echo "make sync       -> build and upload to socialrunclubs.de"
	@echo "make run-remote -> sync & run remote script"

.bin/generate-linux: cmd/generate/main.go go.mod internal/utils/*.go internal/app/*.go templates/*.html templates/parts/*.html
	mkdir -p .bin
	GOOS=linux GOARCH=amd64 go build -o .bin/generate-linux cmd/generate/main.go

.phnoy: check
check:
	go test -v ./...
	go vet ./...

.phony: get-images
get-images:
	go run cmd/get_images/main.go -config local.json

.phony: run-local
run-local:
	rm -rf .out
	go run cmd/generate/main.go -config local.json

.repo/.git/config:
	git clone https://github.com/flopp/socialrunclubs-de.git .repo

.phony: backup
backup:
	@mkdir -p backup-data
	go run cmd/generate/main.go -config local.json -backup backup-data/$(shell date +%Y-%m-%d).ods

.phony: sync
sync: .repo/.git/config .bin/generate-linux
	(cd .repo && git pull --quiet)
	rsync -a production.json scripts/cronjob.sh .bin/generate-linux $(SERVER):$(TARGET_DIR)/
	rsync -a .images/ $(SERVER):$(TARGET_DIR)/images
	rsync -a .repo/ $(SERVER):$(TARGET_DIR)/repo
	ssh $(SERVER) chmod +x $(TARGET_DIR)/cronjob.sh $(TARGET_DIR)/generate-linux

.phony: run-remote
run-remote: sync
	ssh $(SERVER) $(TARGET_DIR)/cronjob.sh
