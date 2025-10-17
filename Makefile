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

# Binary name
BINARY_WINDOWS_NAME=ProcGuardSvc.exe

# Build flags
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all build build-frontend run fmt clean lint install

all: build

build: build-frontend
	mkdir -p build/cache build/bin
	@echo "Generating Windows resources..."
	go generate ./...
	@echo "Building ProcGuardSvc.exe for windows..."
	GOOS=windows $(GO_BUILD) -ldflags="-w -H windowsgui -X main.version=$(VERSION)" -o build/bin/$(BINARY_WINDOWS_NAME) .

build-frontend:
	@echo "Building frontend..."
	cd gui && npm install && npm run build

run:
	$(GO_RUN) . --

fmt:
	@echo "Formatting code..."
	$(GO_FMT) ./...
	cd gui && npm run format

lint:
	GOOS=windows golangci-lint run

clean:
	@echo "Cleaning..."
	$(GO_CLEAN)
	rm -rf build/cache
	rm -rf build/bin
	rm -rf gui/frontend/dist

install:
	@echo "Installing $(BINARY_NAME) to $(shell $(GO_CMD) env GOPATH)/bin..."
	$(GO_INSTALL) .
