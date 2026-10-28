GO ?= go
PKGS := ./...
ROOT_DIR := $(CURDIR)
GOCACHE ?= $(ROOT_DIR)/.cache/go-build
GOMODCACHE ?= $(ROOT_DIR)/.cache/go-mod
GOLANGCI_LINT_CACHE ?= $(ROOT_DIR)/.cache/golangci-lint
BENCHSTAT_VERSION := v0.0.0-20261009192801-be2c69fb417e
BENCHSTAT_BIN := $(ROOT_DIR)/.cache/bin/benchstat-$(BENCHSTAT_VERSION)
BENCHSTAT := $(BENCHSTAT_BIN)/benchstat
export GOCACHE
export GOMODCACHE
export GOLANGCI_LINT_CACHE
unexport GOROOT

.PHONY: help doctor fmt test test-race test-integration bench bench-compare lint tidy build verify

help:
	@echo "Available targets:"
	@echo "  doctor           - Validate local Go toolchain setup"
	@echo "  fmt              - Format Go code"
	@echo "  test             - Run unit tests"
	@echo "  test-race        - Run race-enabled tests"
	@echo "  test-integration - Run integration tests"
	@echo "  bench            - Run benchmarks ten times with allocation stats"
	@echo "  bench-compare    - Compare two bench outputs: OLD=before.txt NEW=after.txt"
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

bench:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) test -run '^$$' -bench . -benchmem -count=10 $(PKGS)

$(BENCHSTAT):
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	GOBIN="$(BENCHSTAT_BIN)" $(GO) install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_VERSION)

bench-compare: $(BENCHSTAT)
	@test -n "$(OLD)" -a -n "$(NEW)" || { echo "usage: make bench-compare OLD=before.txt NEW=after.txt"; exit 1; }
	$(BENCHSTAT) "$(OLD)" "$(NEW)"

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
