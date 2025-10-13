# Makefile for procguard-wails

# Use git describe to get a version string.
VERSION ?= $(shell git describe --tags --always --dirty --first-parent 2>/dev/null || echo "dev")

# Go parameters
GO_CMD=go
GO_FMT=$(GO_CMD) fmt
GO_TEST=$(GO_CMD) test

# Binary name
BINARY_WINDOWS_NAME=procguard-wails.exe

.PHONY: all build run fmt clean test

all: build

build:
	@echo "Building Wails application for windows/amd64..."
	wails build --platform windows/amd64 -ldflags="-X main.version=$(VERSION)"

fmt:
	@echo "Formatting code..."
	$(GO_FMT) ./...
	cd frontend && npm run format

test:
	$(GO_TEST) ./...

clean:
	@echo "Cleaning..."
	rm -rf build/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules
