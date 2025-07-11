.DEFAULT_GOAL := build

.PHONY:fmt vet build test integration
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test -v ./...

integration: vet
	go test -v -count=1 --tags=integration ./...

build: vet
	go build ./runner-manager/cfd
