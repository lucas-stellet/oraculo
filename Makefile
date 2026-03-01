BINARY  := oraculo
BUILD   := ./cmd/oraculo
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install test vet clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

test:
	go test ./... -count=1

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
