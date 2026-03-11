BINARY  := oraculo
BUILD   := ./apps/backend/cmd/oraculo
PREFIX  ?= $(HOME)/.local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
FRONTEND_DIR := ./apps/frontend

.PHONY: build build-frontend build-backend rebuild install test vet clean web-dev cross-compile

build-frontend:
	cd $(FRONTEND_DIR) && bun run build

build-backend:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

build: build-backend

rebuild: build-frontend build-backend

install: build
	install -m 755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

test:
	go test -v -count=1 ./apps/backend/...

vet:
	go vet ./apps/backend/...

clean:
	rm -f $(BINARY)
	rm -rf $(FRONTEND_DIR)/out
	rm -rf npm/cli-*/bin/

web-dev:
	cd $(FRONTEND_DIR) && bun run dev

# Cross-compile for npm distribution (local testing)
cross-compile:
	@mkdir -p npm/cli-darwin-arm64/bin npm/cli-darwin-x64/bin npm/cli-linux-x64/bin npm/cli-linux-arm64/bin
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o npm/cli-darwin-arm64/bin/oraculo  $(BUILD)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o npm/cli-darwin-x64/bin/oraculo    $(BUILD)
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o npm/cli-linux-x64/bin/oraculo     $(BUILD)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o npm/cli-linux-arm64/bin/oraculo   $(BUILD)
