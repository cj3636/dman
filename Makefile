GO ?= go

.PHONY: build run test fmt vet
build:
	$(GO) build -o bin/dman ./cmd/dman
run: build
	./bin/dman
fmt:
	$(GO) fmt ./...
vet:
	$(GO) vet ./...
test:
	$(GO) test ./...
