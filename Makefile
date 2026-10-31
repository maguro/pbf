GO ?= go
PKGS := ./...
ROOT_DIR := $(CURDIR)
GOCACHE ?= $(ROOT_DIR)/.cache/go-build
GOMODCACHE ?= $(ROOT_DIR)/.cache/go-mod
GOLANGCI_LINT_CACHE ?= $(ROOT_DIR)/.cache/golangci-lint
BENCHSTAT_VERSION := v0.0.0-20261009192801-be2c69fb417e
BENCHSTAT_BIN := $(ROOT_DIR)/.cache/bin/benchstat-$(BENCHSTAT_VERSION)
BENCHSTAT := $(BENCHSTAT_BIN)/benchstat
BASE ?= HEAD
BENCH_COUNT ?= 10
GOLANGCI_LINT_VERSION := v2.15.0
GOLANGCI_LINT_BIN := $(ROOT_DIR)/.cache/bin/golangci-lint-$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT := $(GOLANGCI_LINT_BIN)/golangci-lint
GOVULNCHECK_VERSION := v1.9.0
export GOCACHE
export GOMODCACHE
export GOLANGCI_LINT_CACHE
unexport GOROOT

.PHONY: help doctor fmt check-fmt test test-integration coverage bench bench-compare lint vuln tidy build check verify

help:
	@echo "Available targets:"
	@echo "  doctor           - Validate local Go toolchain setup"
	@echo "  fmt              - Format Go code"
	@echo "  check-fmt        - Report formatting changes without writing them"
	@echo "  test             - Run unit tests with the race detector"
	@echo "  test-integration - Run integration tests"
	@echo "  coverage         - Write coverage.out"
	@echo "  bench            - Run benchmarks ten times with allocation stats"
	@echo "  bench-compare    - Compare benchmarks of BASE (default HEAD) with the working tree"
	@echo "  lint             - Run golangci-lint"
	@echo "  vuln             - Check for reachable vulnerabilities with govulncheck"
	@echo "  tidy             - Tidy module dependencies"
	@echo "  build            - Build all packages"
	@echo "  check            - Run format check, tests, and lint"
	@echo "  verify           - Alias for check"

doctor:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOLANGCI_LINT_CACHE)"
	@$(GO) version
	@$(GO) env GOROOT GOVERSION
	@v="$$($(GO) env GOVERSION)"; r="$$(basename "$$($(GO) env GOROOT)")"; \
	  case "$$r" in \
	    go*) [ "$$v" = "$$r" ] || { echo "toolchain mismatch: GOVERSION=$$v, GOROOT=$$r"; exit 1; } ;; \
	  esac

fmt: $(GOLANGCI_LINT)
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOLANGCI_LINT_CACHE)"
	$(GOLANGCI_LINT) fmt

check-fmt: $(GOLANGCI_LINT)
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOLANGCI_LINT_CACHE)"
	$(GOLANGCI_LINT) fmt --diff

test:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) test -v -race -timeout 30s $(PKGS)

test-integration:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) test -v -tags=integration $(PKGS)

coverage:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) test -v -covermode=count -coverprofile=coverage.out -timeout 30s $(PKGS)

bench:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) test -run '^$$' -bench . -benchmem -count=$(BENCH_COUNT) $(PKGS)

$(BENCHSTAT):
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	GOBIN="$(BENCHSTAT_BIN)" $(GO) install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_VERSION)

bench-compare: $(BENCHSTAT)
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	scripts/bench-compare.sh "$(BASE)" "$(BENCH_COUNT)" "$(BENCHSTAT)"

$(GOLANGCI_LINT):
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	GOBIN="$(GOLANGCI_LINT_BIN)" $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint: $(GOLANGCI_LINT)
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)" "$(GOLANGCI_LINT_CACHE)"
	$(GOLANGCI_LINT) run

vuln:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) $(PKGS)

tidy:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) mod tidy

build:
	@mkdir -p "$(GOCACHE)" "$(GOMODCACHE)"
	$(GO) build $(PKGS)

check:
	$(MAKE) doctor
	$(MAKE) check-fmt
	$(MAKE) test
	$(MAKE) test-integration
	$(MAKE) lint

verify: check
