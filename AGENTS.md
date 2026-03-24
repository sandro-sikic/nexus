# AGENTS.md

This file provides essential context for AI agents working on the Nexus project.

## Project Overview

**Nexus** is a CLI tool that transforms scattered terminal commands into a clean, interactive Terminal User Interface (TUI). It consists of:
- **Go CLI Application**: Main TUI tool for running commands
- **React Website**: Promotional/documentation site built with Vite

## Quick Commands

### Go CLI (root directory)
```bash
# Development
make run                    # Run directly with go run
make run-config CONFIG=path # Run with custom config file

# Building
make build                  # Build for current platform
make build-all             # Cross-compile for all platforms
make build-linux           # Build Linux binaries only
make build-darwin          # Build macOS binaries only
make build-windows         # Build Windows binary only

# Testing & Quality
make test                  # Run all tests with race detection
make test-verbose          # Run tests with verbose output
make test-cover            # Generate HTML coverage report
make test-cover-summary    # Print per-package coverage summary
make vet                   # Run go vet
make lint                  # Run golangci-lint
make check                 # Run vet + tests (CI-friendly)

# Maintenance
make tidy                  # Tidy and verify go.mod/go.sum
make deps                  # Download all dependencies
make clean                 # Remove dist/ directory
make help                  # Show all available commands
```

### Web (web/ directory)
```bash
cd web
npm run dev               # Start development server
npm run build             # Build production site
```

## Project Structure

```
nexus/
├── .github/workflows/      # CI/CD workflows
│   ├── build.yml          # Build and release binaries
│   └── deploy.yml         # Deploy website to GitHub Pages
├── config/                # YAML configuration handling
│   ├── config.go          # Config structs and parser
│   └── config_test.go     # Tests
├── dist/                  # Build output (cross-platform binaries)
├── runner/                # Command execution engine
│   ├── runner.go          # Stream, handoff, background execution
│   ├── runner_test.go     # Tests
│   ├── cmd_unix.go        # Unix-specific command handling
│   └── cmd_windows.go     # Windows-specific command handling
├── ui/                    # TUI components (Bubble Tea)
│   ├── app.go             # Main application model
│   ├── fuzzy.go           # Fuzzy search + grouped list view
│   ├── group.go           # Grouped command display
│   ├── list.go            # List component
│   ├── output.go          # Output streaming view
│   ├── styles.go          # Lipgloss styling
│   ├── wizard.go          # Setup/edit wizard
│   └── *_test.go          # Test files
├── web/                   # React website
│   ├── src/
│   │   ├── app/
│   │   │   ├── App.tsx    # Main app component
│   │   │   └── components/# Hero, Features, Wizard, etc.
│   │   ├── styles/
│   │   └── main.tsx       # Entry point
│   ├── package.json
│   └── vite.config.ts
├── main.go                # CLI entry point
├── Makefile               # Build automation
├── nexus.yaml             # Example configuration file
└── README.md              # User documentation
```

## Technology Stack

### Go CLI
- **Go 1.24+**
- **Bubble Tea**: TUI framework (The Elm Architecture for Go)
- **Lipgloss**: Styling library for terminal apps
- **yaml.v3**: YAML parsing for configuration

### Web
- **React 18.3+** with TypeScript
- **Vite 6.3+** for building
- **Tailwind CSS 4.1+** for styling
- **MUI Material 7.3+** for components
- **Framer Motion** for animations
- **Radix UI** primitives

## Key Conventions

### Go Code
- Follow standard Go conventions
- Use `go fmt` for formatting
- Run `make lint` before committing
- Keep functions small and testable
- Place tests in `*_test.go` files alongside source

### Configuration
- Commands are defined in YAML files (default: `nexus.yaml`)
- Three run modes: `stream`, `handoff`, `background`
- Support for grouped commands and multi-step execution

### UI Components
- Use Lipgloss for consistent terminal styling
- Bubble Tea model-update-view pattern
- Separate concerns: UI in `ui/`, logic in `runner/` and `config/`

### Web
- Dark theme (follow existing color scheme)
- Use Tailwind CSS for styling
- Components in `web/src/app/components/`
- Base path is `/nexus/` for GitHub Pages deployment

## CI/CD

### Automated Builds
- Triggers on push to `main` (excluding `web/**` and `*.md` changes)
- Cross-compiles for: Linux (amd64/arm64), macOS (Intel/Apple Silicon), Windows
- Creates/updates a "latest" GitHub release with all binaries

### Website Deployment
- Triggers on changes to `web/**`
- Builds with Vite and deploys to GitHub Pages

## Testing

Always run quality checks before committing:
```bash
make check      # Run go vet and tests
make lint       # Run golangci-lint
```

For comprehensive testing:
```bash
make test-cover # Generate and view coverage report
```

## Dependencies

### Adding Go Dependencies
```bash
go get package@version
make tidy
```

### Adding Web Dependencies
```bash
cd web
npm install package
```

## Important Files

- `nexus.yaml`: Example configuration with inline documentation
- `Makefile`: All build and development commands
- `go.mod`: Go module definition (Go 1.24.2)
- `web/vite.config.ts`: Vite configuration with path aliasing

## Cross-Platform Notes

The CLI supports:
- Windows (amd64)
- macOS (Intel amd64 and Apple Silicon arm64)
- Linux (amd64 and arm64)

Platform-specific code is in:
- `runner/cmd_unix.go`
- `runner/cmd_windows.go`

## License

MIT License (Copyright 2026 Sandro Šikić)
