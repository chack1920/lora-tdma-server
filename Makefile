.PHONY: build clean test package serve update-vendor api statics
PKGS := $(shell go list ./... | grep -v /vendor/ | grep -v lora-tdma-server/api | grep -v /migrations | grep -v /static)
VERSION := $(shell git describe --always |sed -e "s/^v//")

build: statics
	@echo "Compiling source"
	@mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/lora-tdma-server cmd/lora-tdma-server/main.go

statics:
	@echo "Generating static files"
	@go generate cmd/lora-tdma-server/main.go

requirements:
	dep ensure -v
