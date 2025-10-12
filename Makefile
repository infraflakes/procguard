# Makefile for procguard-cli

# Use git describe to get a version string.
# Example: v1.0.0-3-g1234567
# Fallback to 'dev' if not in a git repository.
VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

# Go parameters
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_RUN=$(GO_CMD) run
GO_FMT=$(GO_CMD) fmt
GO_CLEAN=$(GO_CMD) clean
GO_INSTALL=$(GO_CMD) install
GO_TEST=$(GO_CMD) test

# Binary name
BINARY_WINDOWS_NAME=ProcGuardSvc.exe

# Build flags
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all build run fmt clean test install

all: build

build:
	@echo "Generating Windows resources..."
	go generate ./...
	@echo "Building ProcGuardSvc.exe for windows..."
	GOOS=windows $(GO_BUILD) -ldflags="-H windowsgui -X main.version=$(VERSION)" -o build/bin/$(BINARY_WINDOWS_NAME) .

run:
	$(GO_RUN) . --

fmt:
	@echo "Formatting code..."
	$(GO_FMT) ./...

test:
	$(GO_TEST) ./...

clean:
	@echo "Cleaning..."
	$(GO_CLEAN)
	rm -rf build/cache/rsrc.syso
	rm -rf build/bin/$(BINARY_WINDOWS_NAME)

install:
	@echo "Installing $(BINARY_NAME) to $(shell $(GO_CMD) env GOPATH)/bin..."
	$(GO_INSTALL) .
