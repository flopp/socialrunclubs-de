# Coding agent instructions for the socialrunclubs.de project

## General architecture
- This is a static site generator
    - Reads data from Google Sheets
    - Generates static HTML files that are later deployed to a webserver

## Dev environment
- standard go project
- run tests: `go test ./...`
- lint code: `go vet ./...`
- run static site generator: `make run-local`; this renders files into the `.out` folder
- run backup: `make backup`; this download the current Google Sheets data and stores it in the `backup-data` folder

## File structure
- main go programs in `cmd/generate/main.go`, there are helper programs in `cmd/backup/main.go` and `cmd/get_images/main.go` 
- local go package data in `internal`
- HTML templates (go template format): `templates` folder, sub-templates/components in `templates/parts`
- static files (images, css, js) in `static`
