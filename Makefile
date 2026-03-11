BINARY       := oraculo
BUILD        := ./apps/backend/cmd/oraculo
PREFIX       ?= $(HOME)/.local
VERSION      := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS      := -s -w -X main.version=$(VERSION)
FRONTEND_DIR := ./apps/frontend
DESKTOP_APP  := Oraculo.app
DESKTOP_BIN  := apps/desktop/bin/oraculo-desktop

.PHONY: build build-frontend build-backend build-desktop rebuild install install-cli install-desktop test vet clean web-dev cross-compile

build-frontend:
	cd $(FRONTEND_DIR) && bun run build

build-backend:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

build: build-backend

build-desktop: build-backend
	mkdir -p apps/desktop/bin
	cp $(BINARY) apps/desktop/bin/oraculo 2>/dev/null || true
	cd apps/desktop && task build

rebuild: build-frontend build-backend

install: install-cli install-desktop

install-cli: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

install-desktop: build-desktop
	@echo "Packaging $(DESKTOP_APP)..."
	@rm -rf /tmp/$(DESKTOP_APP)
	@mkdir -p /tmp/$(DESKTOP_APP)/Contents/MacOS
	@mkdir -p /tmp/$(DESKTOP_APP)/Contents/Resources
	@cp $(DESKTOP_BIN) /tmp/$(DESKTOP_APP)/Contents/MacOS/oraculo-desktop
	@cp apps/desktop/build/Info.plist /tmp/$(DESKTOP_APP)/Contents/Info.plist
	@cp apps/desktop/build/icons.icns /tmp/$(DESKTOP_APP)/Contents/Resources/icons.icns
	@rm -rf /Applications/$(DESKTOP_APP)
	@cp -r /tmp/$(DESKTOP_APP) /Applications/$(DESKTOP_APP)
	@echo "Installed /Applications/$(DESKTOP_APP)"

test:
	go test -v -count=1 ./apps/backend/...

vet:
	go vet ./apps/backend/...

clean:
	rm -f $(BINARY)
	rm -rf $(FRONTEND_DIR)/out
	rm -rf npm/cli-*/bin/
	rm -rf apps/desktop/frontend/dist

web-dev:
	cd $(FRONTEND_DIR) && bun run dev

# Cross-compile for npm distribution (local testing)
cross-compile:
	@mkdir -p npm/cli-darwin-arm64/bin npm/cli-darwin-x64/bin npm/cli-linux-x64/bin npm/cli-linux-arm64/bin
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o npm/cli-darwin-arm64/bin/oraculo  $(BUILD)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o npm/cli-darwin-x64/bin/oraculo    $(BUILD)
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o npm/cli-linux-x64/bin/oraculo     $(BUILD)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o npm/cli-linux-arm64/bin/oraculo   $(BUILD)
