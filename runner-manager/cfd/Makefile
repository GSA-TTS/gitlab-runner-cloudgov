.DEFAULT_GOAL := build

.PHONY:fmt vet build test integration
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test ./...

integration: vet
	go test -count=1 --tags=integration ./...

build: vet
	go build
