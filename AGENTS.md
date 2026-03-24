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
│   ├── build.yml          # Build and release binaries (cross-platform)
│   └── deploy.yml         # Deploy website to GitHub Pages
├── config/                # Configuration package (YAML handling)
│   ├── config.go          # Config structures, parsing, Load/Write functions
│   └── config_test.go     # Configuration tests (272 lines)
├── runner/                # Command execution engine
│   ├── runner.go          # Stream, handoff, background execution (431 lines)
│   ├── runner_test.go     # Comprehensive runner tests (895 lines)
│   ├── registry.go        # Process registry for cleanup (67 lines)
│   ├── cmd_unix.go        # Unix-specific command handling (sh -c execution)
│   └── cmd_windows.go     # Windows-specific command handling (cmd.exe /C)
├── ui/                    # TUI components (Bubble Tea framework)
│   ├── app.go             # Main application state machine & signal handling
│   ├── app_test.go        # App tests (685 lines)
│   ├── fuzzy.go           # Fuzzy search + grouped list view
│   ├── fuzzy_test.go      # Fuzzy search tests
│   ├── output.go          # Real-time output streaming view with spinner
│   ├── output_test.go     # Output component tests
│   ├── wizard.go          # Interactive setup/edit wizard (CREATE/EDIT modes)
│   ├── wizard_test.go     # Wizard tests
│   ├── list.go            # List navigation component
│   ├── list_test.go       # List component tests
│   ├── group.go           # Grouped command display with headers
│   ├── group_test.go      # Group rendering tests
│   ├── styles.go          # Lipgloss styling definitions
│   └── *_test.go          # 8 test files total (1852 lines of tests)
├── web/                   # React promotional website
│   ├── src/
│   │   ├── app/
│   │   │   ├── App.tsx    # Main React component
│   │   │   └── components/# 9 presentation components (Hero, Features, etc.)
│   │   ├── styles/        # CSS files (index, tailwind, theme)
│   │   └── main.tsx       # React 18 entry point
│   ├── vite.config.ts     # Vite config (base: /nexus/, plugins)
│   ├── postcss.config.mjs # PostCSS configuration
│   ├── package.json       # Web dependencies
│   └── package-lock.json
├── main.go                # CLI entry point (105 lines) - flag parsing & flow control
├── Makefile               # Build automation (177 lines)
├── nexus.yaml             # Example configuration file with documentation
├── go.mod                 # Go module definition (Go 1.24.2)
├── go.sum                 # Dependency checksums
├── README.md              # User documentation
├── LICENSE                # MIT License
└── .gitignore
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
- **Material-UI (MUI) 7.3+** for components
- **Radix UI** for accessible primitives
- **Framer Motion** (`motion` package) for animations
- **React Router 7.13+** for navigation
- **Recharts 2.15+** for charts
- **Sonner** for toast notifications

## Key Conventions

### Go Code
- Follow standard Go conventions
- Use `go fmt` for formatting
- Run `make lint` before committing
- Keep functions small and testable
- Place tests in `*_test.go` files alongside source

### Go Application Architecture

#### Entry Point (`main.go` - 105 lines)
- CLI flag parsing: `-config` (default: `nexus.yaml`), `-wizard` (edit mode)
- Signal handling for graceful shutdown (SIGINT, SIGTERM)
- Flow control: setup wizard → fuzzy search menu → task execution

#### Config Package (`config/` - ~370 lines total)
- **Structures**: `Config`, `Task`, `Action`
- **YAML parsing**: Uses `gopkg.in/yaml.v3`
- **Task methods**: `AllCommands()`, `HasBackgroundActions()`, `HasHandoff()`, `LastAction()`
- **Functions**: `Load()` reads YAML, `Write()` saves configuration
- **Default title**: "Nexus"

#### Runner Package (`runner/` - ~1400 lines total)
Core command execution engine with three execution modes:

1. **Stream Mode** (default)
   - Executes actions sequentially in foreground
   - Streams output to TUI in real-time
   - Blocks until all actions complete
   - Used for typical command workflows

2. **Background Mode**
   - Non-blocking execution (starts and returns immediately)
   - Spawns all actions in background process group
   - Output accessible via background process viewer
   - Used for long-running services (dev servers, watchers)

3. **Handoff Mode**
   - Last action assumes raw terminal control
   - Child process becomes interactive terminal session
   - Ctrl+C handled by child process (native behavior)
   - Used for REPL tools, editors, shells

**Key Components:**
- **`runner.go`** (431 lines): Stream, background, and handoff execution
  - `BackgroundProc` struct: Tracks background process metadata
  - Output line buffering (1 MB per line capacity)
  - Process group management for child process cleanup
  
- **`registry.go`** (67 lines): Process lifecycle management
  - Thread-safe `ProcessRegistry` for all spawned processes
  - `KillAll()` method: Cleans up all processes on application exit
  - Called on SIGINT/SIGTERM to prevent orphaned processes
  
- **`cmd_unix.go`**: Unix/Linux command execution
  - Uses `sh -c` for shell command execution
  - Creates process group with `setpgid()` for group signals
  - SIGTERM → 100ms sleep → SIGKILL cleanup sequence
  
- **`cmd_windows.go`**: Windows command execution
  - Uses `cmd.exe /C` for shell commands
  - `CREATE_NEW_PROCESS_GROUP` flag for process tree isolation
  - `taskkill /F /T` for terminating process trees

**Testing**: 895 lines of comprehensive tests covering all modes, error handling, and process cleanup

#### UI Package (`ui/` - Bubble Tea TUI Framework)
State machine with three states: menu (fuzzy search), output (streaming), background (process viewer)

- **`app.go`** (Main state machine)
  - Three application states: `stateMenu`, `stateOutput`, `stateBG`
  - Global quit shortcuts: `q`, `Ctrl+C`
  - Signal handlers integrated with process registry
  - Message passing for state transitions
  
- **`fuzzy.go`** (Search + grouping)
  - Fuzzy filter with real-time matching
  - Dynamic grouping by `Task.Group` field
  - Scroll management with viewport
  
- **`output.go`** (Streaming view)
  - Real-time log line streaming to TUI
  - Spinner animation during execution
  - Optional stderr highlighting
  
- **`wizard.go`** (Interactive setup)
  - CREATE mode: Build new configuration from scratch
  - EDIT mode: Modify existing configuration
  - Multi-step TUI form for task management
  
- **`list.go`, `group.go`, `styles.go`**: Supporting components
  - List navigation and selection
  - Group header rendering
  - Lipgloss styling (colors, layout, spacing)

**Test Coverage**: 8 test files with 1852 total lines
- Comprehensive state machine testing
- UI component rendering and interaction
- Edge cases and error conditions

### Configuration
- Commands are defined in YAML files (default: `nexus.yaml`)
- **Task structure**: Each task contains `name`, `description`, `actions[]` array, optional `dir`, `group`, and `handoff`
- **Actions format**: Each action has `command` string and optional `background: true/false` flag
- **Execution modes**:
  - **stream** (default): Runs actions sequentially, streams output in TUI, blocking
  - **background**: Runs all actions in background, returns immediately, accessible via background process viewer
  - **handoff**: Last action takes over raw terminal control (Ctrl+C works as native terminal)
- Support for grouped commands with multi-step execution
- **Example structure**:
  ```yaml
  title: "My Project"
  tasks:
    - name: "Start Dev Server"
      description: "Run development server"
      actions:
        - command: "npm run dev"
      group: "Development"
    - name: "Build & Test"
      description: "Build and run tests"
      actions:
        - command: "npm run build"
        - command: "npm test"
      background: true
    - name: "Interactive Shell"
      description: "Open shell in project directory"
      actions:
        - command: "bash"
      handoff: true
      dir: "/path/to/project"
  ```

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
