GO ?= go

VERSION ?= dev
COMMIT  ?= unknown
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -X git.tyss.io/cj3636/dman/internal/buildinfo.Version=$(VERSION) -X git.tyss.io/cj3636/dman/internal/buildinfo.Commit=$(COMMIT) -X git.tyss.io/cj3636/dman/internal/buildinfo.BuildTime=$(BUILD_TIME)

.PHONY: build run test fmt vet test-race coverage
build:
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/dman ./cmd/dman
run: build
	./bin/dman
fmt:
	$(GO) fmt ./...
vet:
	$(GO) vet ./...
test:
	$(GO) test ./...
test-race:
	$(GO) test -race ./...
coverage:
	$(GO) test -coverprofile=coverage.out ./... && $(GO) tool cover -func=coverage.out | grep total:
.PHONY: coverage-threshold
coverage-threshold:
	$(GO) test -coverprofile=coverage.out ./... >/dev/null
	@total=$$(go tool cover -func=coverage.out | grep total: | awk '{gsub(/%/,"",$$3); split($$3,a,"."); print a[1]}'); \
	thresh=70; \
	if [ $$total -lt $$thresh ]; then \
		echo "Coverage $$total% below threshold $$thresh%"; exit 1; \
	else \
		echo "Coverage $$total% OK (threshold $$thresh%)"; \
	fi
