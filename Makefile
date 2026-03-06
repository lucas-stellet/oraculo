BINARY  := oraculo
BUILD   := ./cmd/oraculo
PREFIX  ?= $(HOME)/.local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
DASHBOARD_DIR := ./apps/dashboard

.PHONY: build install test vet clean web-dev

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

install: build
	install -m 755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

test:
	go test -v -count=1 ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

web-dev:
	cd $(DASHBOARD_DIR) && bun run dev
