GO ?= go
PKGS := ./...
ROOT_DIR := $(CURDIR)
GOCACHE ?= $(ROOT_DIR)/.cache/go-build
GOMODCACHE ?= $(ROOT_DIR)/.cache/go-mod
GOLANGCI_LINT_CACHE ?= $(ROOT_DIR)/.cache/golangci-lint
export GOCACHE
export GOMODCACHE
export GOLANGCI_LINT_CACHE
unexport GOROOT

.PHONY: help doctor fmt test test-race test-integration lint tidy build verify

help:
	@echo "Available targets:"
	@echo "  doctor           - Validate local Go toolchain setup"
	@echo "  fmt              - Format Go code"
	@echo "  test             - Run unit tests"
	@echo "  test-race        - Run race-enabled tests"
	@echo "  test-integration - Run integration tests"
	@echo "  lint             - Run golangci-lint"
	@echo "  tidy             - Tidy module dependencies"
	@echo "  build            - Build all packages"
	@echo "  verify           - Run fmt check, tests, lint, and build"

doctor:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOLANGCI_LINT_CACHE)"
	@$(GO) version
	@$(GO) env GOROOT GOVERSION
	@v="$$($(GO) env GOVERSION)"; r="$$(basename "$$($(GO) env GOROOT)")"; \
	  case "$$r" in \
	    go*) [ "$$v" = "$$r" ] || { echo "toolchain mismatch: GOVERSION=$$v, GOROOT=$$r"; exit 1; } ;; \
	  esac

fmt:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) fmt $(PKGS)

test:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) test -v -timeout 30s $(PKGS)

test-race:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) test -v -race -timeout 30s $(PKGS)

test-integration:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) test -v -tags=integration $(PKGS)

lint:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOLANGCI_LINT_CACHE)"
	golangci-lint run

tidy:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) mod tidy

build:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) build $(PKGS)

verify:
	$(MAKE) doctor
	@echo "Checking formatting..."
	@test -z "$$($(GO) fmt $(PKGS))" || (echo "Formatting changes required. Run 'make fmt'." && exit 1)
	$(MAKE) test
	$(MAKE) test-integration
	$(MAKE) lint
	$(MAKE) build
