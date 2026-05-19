BINARY=system-hub
BINARY_WIN=system-hub.exe
GO=go
BUILD_DIR=build

.PHONY: all build clean test lint run install

all: build

build:
	$(GO) build -o $(BINARY) ./cmd/system-hub/

run:
	$(GO) run ./cmd/system-hub/

test:
	$(GO) test ./... -v

lint:
	$(GO) vet ./...

clean:
	rm -rf $(BUILD_DIR) $(BINARY) $(BINARY_WIN)
	$(GO) clean

install: build
	install -d $(DESTDIR)/usr/local/bin
	install -m 755 $(BINARY) $(DESTDIR)/usr/local/bin/$(BINARY)

# Cross-compilation
linux-amd64:
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/system-hub/

linux-arm64:
	GOOS=linux GOARCH=arm64 $(GO) build -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/system-hub/

windows-amd64:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_WIN) ./cmd/system-hub/

cross: linux-amd64 linux-arm64 windows-amd64
