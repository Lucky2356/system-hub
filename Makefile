BINARY=system-hub
BINARY_WIN=system-hub.exe
GO=go
BUILD_DIR=build
DIST_DIR=dist

# Version is injected into the binary. Override with: make build VERSION=v1.2.3
VERSION ?= dev
VERSION_PKG=github.com/Lucky2356/system-hub/internal/version.Version
LDFLAGS=-s -w -X $(VERSION_PKG)=$(VERSION)

.PHONY: all build clean test lint run install vendor \
        linux-amd64 linux-arm64 windows-amd64 cross \
        package-deb package-rpm packages

all: build

build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/system-hub/

run:
	$(GO) run ./cmd/system-hub/

test:
	$(GO) test ./... -count=1

lint:
	$(GO) vet ./...

vendor:
	$(GO) mod vendor

clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR) $(BINARY) $(BINARY_WIN)
	$(GO) clean

install: build
	install -d $(DESTDIR)/usr/local/bin
	install -m 755 $(BINARY) $(DESTDIR)/usr/local/bin/$(BINARY)

# --- Cross-compilation --------------------------------------------------------
# NOTE: Fyne requires CGO. Cross-compiling therefore needs a matching C
# toolchain (e.g. mingw-w64 for Windows). The plain targets below assume you
# run them on the target OS, or that CC/CXX point at a cross toolchain.

linux-amd64:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/system-hub/

linux-arm64:
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/system-hub/

# Build on Windows (native), or with CC=x86_64-w64-mingw32-gcc for cross builds.
windows-amd64:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags "$(LDFLAGS) -H=windowsgui" -o $(BUILD_DIR)/$(BINARY_WIN) ./cmd/system-hub/

cross: linux-amd64 linux-arm64 windows-amd64

# --- Linux packages (.deb / .rpm) via nfpm ------------------------------------
# Requires: nfpm (https://nfpm.goreleaser.com) and a Linux binary named ./system-hub.
# Example: make packages VERSION=0.1.0
package-deb: build
	mkdir -p $(DIST_DIR)
	VERSION=$(patsubst v%,%,$(VERSION)) GOARCH=amd64 nfpm pkg --packager deb --target $(DIST_DIR)/

package-rpm: build
	mkdir -p $(DIST_DIR)
	VERSION=$(patsubst v%,%,$(VERSION)) GOARCH=amd64 nfpm pkg --packager rpm --target $(DIST_DIR)/

packages: package-deb package-rpm
