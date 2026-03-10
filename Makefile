BINARY  := oraculo
BUILD   := ./apps/backend/cmd/oraculo
PREFIX  ?= $(HOME)/.local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
DASHBOARD_DIR := ./apps/dashboard

.PHONY: build rebuild install test vet clean web-dev

build:
	rm -rf $(DASHBOARD_DIR)/out apps/backend/src/server/dashboard_assets
	cd $(DASHBOARD_DIR) && bun run build
	mkdir -p apps/backend/src/server/dashboard_assets
	cp -r $(DASHBOARD_DIR)/out/* apps/backend/src/server/dashboard_assets/
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

rebuild:
	@# Force full rebuild: dashboard + Go binary
	rm -rf $(DASHBOARD_DIR)/out apps/backend/src/server/dashboard_assets
	cd $(DASHBOARD_DIR) && bun run build
	mkdir -p apps/backend/src/server/dashboard_assets
	cp -r $(DASHBOARD_DIR)/out/* apps/backend/src/server/dashboard_assets/
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

install: build
	install -m 755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

test:
	go test -v -count=1 ./apps/backend/...

vet:
	go vet ./apps/backend/...

clean:
	rm -f $(BINARY)
	rm -rf apps/backend/src/server/dashboard_assets

web-dev:
	cd $(DASHBOARD_DIR) && bun run dev
