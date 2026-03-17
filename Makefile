# ── Config ────────────────────────────────────────────────────────────────────
BINARY := runner
MODULE := runner

# Use = (recursive) instead of := (immediate) so these are evaluated only when
# a build target actually runs, never at parse time.  This prevents hangs when
# git or date are slow / missing on the current PATH.
VERSION    = $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     = $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_TIME = $(shell git log -1 --format=%cI 2>/dev/null || echo unknown)

DIST_DIR   := dist
MAIN_PKG   := .

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
	@mkdir -p $(DIST_DIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)$(shell go env GOEXE) $(MAIN_PKG)
	@echo "Built → $(DIST_DIR)/$(BINARY)$(shell go env GOEXE)"

.PHONY: build-all
build-all: ## Cross-compile for all platforms → dist/
	@mkdir -p $(DIST_DIR)
	@$(foreach PLATFORM,$(PLATFORMS), \
		$(eval OS   := $(word 1,$(subst /, ,$(PLATFORM)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \
		echo "Building $(BINARY) for $(OS)/$(ARCH)…"; \
		GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) \
			-o $(call bin_name,$(PLATFORM)) \
			$(MAIN_PKG); \
	)
	@echo "All binaries written to $(DIST_DIR)/"

.PHONY: build-linux
build-linux: ## Build Linux amd64 + arm64 binaries
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64  $(GOBUILD) -o $(DIST_DIR)/$(BINARY)_linux_amd64  $(MAIN_PKG)
	GOOS=linux GOARCH=arm64  $(GOBUILD) -o $(DIST_DIR)/$(BINARY)_linux_arm64  $(MAIN_PKG)

.PHONY: build-darwin
build-darwin: ## Build macOS amd64 + arm64 (Apple Silicon) binaries
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/$(BINARY)_darwin_amd64 $(MAIN_PKG)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(DIST_DIR)/$(BINARY)_darwin_arm64 $(MAIN_PKG)

.PHONY: build-windows
build-windows: ## Build Windows amd64 binary
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/$(BINARY)_windows_amd64.exe $(MAIN_PKG)

# ── Test ──────────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run all tests
	$(GOTEST) -race ./...

.PHONY: test-verbose
test-verbose: ## Run all tests with verbose output
	$(GOTEST) -race -v ./...

.PHONY: test-cover
test-cover: ## Run tests and open HTML coverage report
	@mkdir -p $(DIST_DIR)
	$(GOTEST) -race -coverprofile=$(DIST_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(DIST_DIR)/coverage.out -o $(DIST_DIR)/coverage.html
	@echo "Coverage report → $(DIST_DIR)/coverage.html"

.PHONY: test-cover-summary
test-cover-summary: ## Print per-package coverage summary
	@mkdir -p $(DIST_DIR)
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
	rm -rf $(DIST_DIR)
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
