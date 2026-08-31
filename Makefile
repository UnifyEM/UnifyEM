.PHONY: all build clean install uninstall test vet fmt lint help

# UnifyEM builds three binaries from one module. Each component's package main
# lives in the directory named here, and the binary it produces is uem-<name>.
COMPONENTS=server cli agent

BUILD_DIR=bin

# Go variables. CGO is disabled so every target cross-compiles from any host.
GO?=CGO_ENABLED=0 go
GOFLAGS?=
LDFLAGS=-ldflags "-s -w"

# The version is NOT injected here. It is the Version/Build const pair in
# common/version.go and is a property of the source, not of the machine that
# compiled it. b9 reads that same file to name the release, so a binary can
# never disagree with the release it ships in.

GOLANGCI_LINT?=golangci-lint

# Installation
INSTALL_PREFIX?=$(HOME)/.local
INSTALL_BIN_DIR=$(INSTALL_PREFIX)/bin

# OS detection
UNAME_S:=$(shell uname -s)
UNAME_M:=$(shell uname -m)

# PLATFORM maps to GOOS and ARCH to GOARCH. Both default to the detected host,
# so a plain `make build` is unchanged; override either on the command line to
# cross-compile:
#   make build PLATFORM=linux   ARCH=arm64
#   make build PLATFORM=windows ARCH=amd64
#   make build PLATFORM=darwin  ARCH=arm64
ifeq ($(UNAME_S),Darwin)
PLATFORM?=darwin
else ifeq ($(UNAME_S),Linux)
PLATFORM?=linux
else
PLATFORM?=$(UNAME_S)
endif

ifeq ($(UNAME_M),x86_64)
ARCH?=amd64
else ifeq ($(UNAME_M),aarch64)
ARCH?=arm64
else ifeq ($(UNAME_M),arm64)
ARCH?=arm64
else
ARCH?=$(UNAME_M)
endif

# Windows executables need the .exe suffix: Windows will not run an
# extensionless binary and jsign will not sign one.
EXE=$(if $(filter windows,$(PLATFORM)),.exe)

BUILD_ENV=GOOS=$(PLATFORM) GOARCH=$(ARCH)

# Default target
all: build

## build: Build uem-server, uem-cli and uem-agent for PLATFORM/ARCH (host by default).
build:
	@mkdir -p $(BUILD_DIR)
	@for c in $(COMPONENTS); do \
		out=$(BUILD_DIR)/uem-$$c-$(PLATFORM)-$(ARCH)$(EXE); \
		echo "Building uem-$$c for $(PLATFORM)/$(ARCH)..."; \
		$(BUILD_ENV) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $$out ./$$c || exit 1; \
		echo "Build complete: $$out"; \
	done

## clean: Remove all build artifacts.
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

## install: Build for the host and install the three binaries into the prefix.
install: build
	@mkdir -p $(INSTALL_BIN_DIR)
	@for c in $(COMPONENTS); do \
		install -m 0755 $(BUILD_DIR)/uem-$$c-$(PLATFORM)-$(ARCH)$(EXE) \
			$(INSTALL_BIN_DIR)/uem-$$c$(EXE) || exit 1; \
		echo "Installed $(INSTALL_BIN_DIR)/uem-$$c$(EXE)"; \
	done

## uninstall: Remove the installed binaries from the prefix.
uninstall:
	@for c in $(COMPONENTS); do \
		rm -f $(INSTALL_BIN_DIR)/uem-$$c$(EXE); \
		echo "Removed $(INSTALL_BIN_DIR)/uem-$$c$(EXE)"; \
	done

## test: Run the test suite.
test:
	$(GO) test ./...

## vet: Run go vet.
vet:
	$(GO) vet ./...

## fmt: Format the source.
fmt:
	gofmt -w .

## lint: Run golangci-lint.
lint:
	$(GOLANGCI_LINT) run ./...

## help: Show this help.
help:
	@echo "UnifyEM Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target] [PLATFORM=<goos>] [ARCH=<goarch>]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sort | awk -F': ' '{printf "  %-20s %s\n", substr($$1, 4), $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make                                       # Build for the host"
	@echo "  make build PLATFORM=windows ARCH=amd64     # Windows x86-64"
	@echo "  make build PLATFORM=linux   ARCH=arm64     # Linux arm64"
	@echo "  make install                               # Install to $(INSTALL_BIN_DIR)"
