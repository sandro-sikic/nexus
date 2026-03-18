# ── Config ────────────────────────────────────────────────────────────────────
BINARY := runner
MODULE := runner

# Use = (recursive) instead of := (immediate) so these are evaluated only when
# a build target actually runs, never at parse time.  This prevents hangs when
# git or date are slow / missing on the current PATH.
VERSION    = $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     = $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_TIME = $(shell git log -1 --format=%cI 2>/dev/null || echo unknown)

DIST_DIR := dist
MAIN_PKG := .

# ── Platform detection ────────────────────────────────────────────────────────
# go env GOOS is the only reliable cross-platform way to detect Windows here.
ifeq ($(shell go env GOOS),windows)
  MKDIR = cmd /c if not exist "$(DIST_DIR)" mkdir "$(DIST_DIR)"
  RM    = cmd /c if exist "$(DIST_DIR)" rmdir /s /q "$(DIST_DIR)"
else
  MKDIR = mkdir -p $(DIST_DIR)
  RM    = rm -rf $(DIST_DIR)
endif

LDFLAGS = -s -w \
	-X '$(MODULE)/internal/version.Version=$(VERSION)' \
	-X '$(MODULE)/internal/version.Commit=$(COMMIT)' \
	-X '$(MODULE)/internal/version.BuildTime=$(BUILD_TIME)'

GO      := go
GOBUILD  = $(GO) build -trimpath -ldflags "$(LDFLAGS)"
GOTEST  := $(GO) test
GOVET   := $(GO) vet

# Cross-compile targets: OS/ARCH pairs
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

# ── Default ───────────────────────────────────────────────────────────────────
.DEFAULT_GOAL := build

# ── Helpers ───────────────────────────────────────────────────────────────────

# binary name with optional .exe suffix
define bin_name
$(DIST_DIR)/$(BINARY)_$(subst /,_,$(1))$(if $(filter windows/%,$(1)),.exe,)
endef

# ── Primary targets ───────────────────────────────────────────────────────────

.PHONY: build
build: ## Build for the current platform → dist/
	@$(MKDIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)$(shell go env GOEXE) $(MAIN_PKG)
	@echo "Built → $(DIST_DIR)/$(BINARY)$(shell go env GOEXE)"

.PHONY: build-all
build-all: ## Cross-compile for all platforms → dist/
	@$(MKDIR)
	$(call build_for,linux,amd64)
	$(call build_for,linux,arm64)
	$(call build_for,darwin,amd64)
	$(call build_for,darwin,arm64)
	$(call build_for,windows,amd64)
	@echo "All binaries written to $(DIST_DIR)/"

# Cross-platform helper for setting GOOS/GOARCH (Windows cmd vs Unix shell)
ifeq ($(OS),Windows_NT)
  define build_for
	@echo "Building $(BINARY) for $(1)/$(2)…" && set "GOOS=$(1)" && set "GOARCH=$(2)" && $(GOBUILD) -o $(call bin_name,$(1)/$(2)) $(MAIN_PKG)
  endef
else
  define build_for
	@echo "Building $(BINARY) for $(1)/$(2)…" && GOOS=$(1) GOARCH=$(2) $(GOBUILD) -o $(call bin_name,$(1)/$(2)) $(MAIN_PKG)
  endef
endif

.PHONY: build-linux
build-linux: ## Build Linux amd64 + arm64 binaries
	@$(MKDIR)
	$(call build_for,linux,amd64)
	$(call build_for,linux,arm64)

.PHONY: build-darwin
build-darwin: ## Build macOS amd64 + arm64 (Apple Silicon) binaries
	@$(MKDIR)
	$(call build_for,darwin,amd64)
	$(call build_for,darwin,arm64)

.PHONY: build-windows
build-windows: ## Build Windows amd64 binary
	@$(MKDIR)
	$(call build_for,windows,amd64)

# ── Test ──────────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run all tests
	$(GOTEST) -race ./...

.PHONY: test-verbose
test-verbose: ## Run all tests with verbose output
	$(GOTEST) -race -v ./...

.PHONY: test-cover
test-cover: ## Run tests and open HTML coverage report
	@$(MKDIR)
	$(GOTEST) -race -coverprofile=$(DIST_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(DIST_DIR)/coverage.out -o $(DIST_DIR)/coverage.html
	@echo "Coverage report → $(DIST_DIR)/coverage.html"

.PHONY: test-cover-summary
test-cover-summary: ## Print per-package coverage summary
	@$(MKDIR)
	$(GOTEST) -race -coverprofile=$(DIST_DIR)/coverage.out ./...
	$(GO) tool cover -func=$(DIST_DIR)/coverage.out

# ── Quality ───────────────────────────────────────────────────────────────────

.PHONY: vet
vet: ## Run go vet
	$(GOVET) ./...

.PHONY: lint
lint: ## Run golangci-lint (must be installed)
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found — install from https://golangci-lint.run"; exit 1; }
	golangci-lint run ./...

.PHONY: check
check: vet test ## Run vet + tests (CI-friendly)

# ── Deps ──────────────────────────────────────────────────────────────────────

.PHONY: tidy
tidy: ## Tidy and verify go.mod / go.sum
	$(GO) mod tidy
	$(GO) mod verify

.PHONY: deps
deps: ## Download all dependencies
	$(GO) mod download

# ── Clean ─────────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove the dist/ directory
	@$(RM)
	@echo "Cleaned $(DIST_DIR)/"

# ── Run (dev shortcut) ────────────────────────────────────────────────────────

.PHONY: run
run: ## Run directly with go run (uses runner.yaml in CWD)
	$(GO) run $(MAIN_PKG)

.PHONY: run-config
run-config: ## Run with a custom config: make run-config CONFIG=path/to/file.yaml
	$(GO) run $(MAIN_PKG) --config $(CONFIG)

# ── Info ──────────────────────────────────────────────────────────────────────

.PHONY: version
version: ## Print resolved version info
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Go version: $(shell $(GO) version)"

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
