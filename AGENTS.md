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
make build-all              # Cross-compile for all platforms
make build-linux            # Build Linux binaries only
make build-darwin           # Build macOS binaries only
make build-windows          # Build Windows binary only

# Testing & Quality
make test                   # Run all tests with race detection
make test-verbose           # Run tests with verbose output
make test-cover             # Generate HTML coverage report
make test-cover-summary     # Print per-package coverage summary
make vet                    # Run go vet
make lint                   # Run golangci-lint
make check                  # Run vet + tests (CI-friendly)

# Maintenance
make tidy                   # Tidy and verify go.mod/go.sum
make deps                   # Download all dependencies
make clean                  # Remove dist/ directory
make version                # Print current version
make help                   # Show all available commands
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
│   ├── config.go          # Config structures, parsing, Load/Write functions (101 lines)
│   └── config_test.go     # Configuration tests (272 lines)
├── runner/                # Command execution engine (~1497 lines total)
│   ├── runner.go          # Stream, handoff, background execution (431 lines)
│   ├── runner_test.go     # Comprehensive runner tests (895 lines)
│   ├── registry.go        # Process registry for cleanup (67 lines)
│   ├── cmd_unix.go        # Unix-specific command handling (58 lines)
│   └── cmd_windows.go     # Windows-specific command handling (46 lines)
├── ui/                    # TUI components (Bubble Tea framework)
│   ├── app.go             # Main application state machine (238 lines)
│   ├── app_test.go        # App tests (685 lines)
│   ├── fuzzy.go           # Fuzzy search + grouped list view (421 lines)
│   ├── fuzzy_test.go      # Fuzzy search tests (455 lines)
│   ├── output.go          # Real-time output streaming + background viewer (314 lines)
│   ├── output_test.go     # Output component tests (508 lines)
│   ├── wizard.go          # Interactive setup/edit wizard (1109 lines)
│   ├── wizard_test.go     # Wizard tests (1544 lines)
│   ├── list.go            # List navigation component (92 lines)
│   ├── list_test.go       # List component tests (265 lines)
│   ├── group.go           # Grouped command display with headers (145 lines)
│   ├── group_test.go      # Group rendering tests (291 lines)
│   └── styles.go          # Lipgloss styling definitions (59 lines)
├── web/                   # React promotional website
│   ├── src/
│   │   ├── app/
│   │   │   ├── App.tsx    # Root layout, renders all 9 section components (30 lines)
│   │   │   └── components/
│   │   │       ├── Hero.tsx           # Animated hero with floating commands (266 lines)
│   │   │       ├── Features.tsx       # 6-feature grid with scroll animations (82 lines)
│   │   │       ├── HowItWorks.tsx     # 3-step flow diagram (75 lines)
│   │   │       ├── Wizard.tsx         # Wizard feature explainer (131 lines)
│   │   │       ├── TUIPreview.tsx     # Mocked terminal TUI preview (116 lines)
│   │   │       ├── CodeExample.tsx    # Syntax-highlighted YAML + copy button (295 lines)
│   │   │       ├── RealWorldExample.tsx # Before/After comparison (113 lines)
│   │   │       ├── CTA.tsx            # Call-to-action section (74 lines)
│   │   │       └── Footer.tsx         # 4-column footer with nav links (145 lines)
│   │   ├── styles/
│   │   │   ├── index.css   # Imports tailwind.css and theme.css
│   │   │   ├── tailwind.css # Tailwind v4 config with source glob
│   │   │   └── theme.css   # Full CSS custom property theme (oklch colors, 181 lines)
│   │   └── main.tsx        # React 18 createRoot entry point (7 lines)
│   ├── index.html          # Vite HTML shell (15 lines)
│   ├── vite.config.ts      # Vite config (base: /nexus/, plugins, @ alias)
│   ├── postcss.config.mjs  # PostCSS configuration (Tailwind v4 handles PostCSS via plugin)
│   ├── package.json        # Web dependencies
│   └── package-lock.json
├── dist/                  # Pre-built binaries (5 targets, git-tracked)
│   ├── nexus_darwin_amd64
│   ├── nexus_darwin_arm64
│   ├── nexus_linux_amd64
│   ├── nexus_linux_arm64
│   └── nexus_windows_amd64.exe
├── main.go                # CLI entry point (105 lines)
├── Makefile               # Build automation (177 lines)
├── nexus.yaml             # Example configuration file with inline documentation (119 lines)
├── go.mod                 # Go module definition (module: nexus, Go 1.24.2)
├── go.sum                 # Dependency checksums
├── README.md              # User documentation (165 lines)
├── LICENSE                # MIT License (Copyright 2026 Sandro Šikić)
└── .gitignore             # Ignores dist/ and node_modules/
```

## Technology Stack

### Go CLI
- **Go 1.24.2**
- **`charmbracelet/bubbletea` v1.3.10**: TUI framework (The Elm Architecture for Go)
- **`charmbracelet/bubbles` v1.0.0**: Reusable TUI components
- **`charmbracelet/lipgloss` v1.1.0**: Terminal styling library
- **`gopkg.in/yaml.v3` v3.0.1**: YAML parsing for configuration

### Web
- **React 18.3.1** with TypeScript (listed as optional peer dependency)
- **Vite 6.3.5** for building
- **`@vitejs/plugin-react` 4.7.0** for React support
- **Tailwind CSS 4.1.12** for styling (via `@tailwindcss/vite` plugin)
- **Material-UI (MUI) 7.3.5** + `@emotion/react` 11.14.0 / `@emotion/styled` 11.14.1
- **Radix UI** (26 primitive packages) for accessible components
- **`motion` 12.23.24** (Framer Motion) for animations
- **React Router 7.13.0** for navigation
- **Recharts 2.15.2** for charts
- **Sonner 2.0.3** for toast notifications
- **lucide-react 0.487.0** for icons
- **`next-themes` 0.4.6** for theme switching
- **`react-hook-form` 7.55.0**, **`cmdk` 1.1.1**, **`vaul` 1.1.2**, and many more UI utilities

## Key Conventions

### Go Code
- Follow standard Go conventions
- Use `go fmt` for formatting
- Run `make lint` before committing (uses `golangci-lint`)
- Keep functions small and testable
- Place tests in `*_test.go` files alongside source

### Go Application Architecture

#### Entry Point (`main.go` - 105 lines)
- CLI flag parsing: `-config` (default: `nexus.yaml`), `-wizard` (edit mode)
- Signal handling for graceful shutdown (SIGINT, SIGTERM) via `ProcessRegistry.KillAll()`
- Flow control: if config missing or `-wizard` passed → run wizard → load config → launch fuzzy TUI → handle handoff if selected task requires it

#### Config Package (`config/config.go` - 101 lines)
- **Structures**:
  - `Action` — `Command string`
  - `Task` — `Name`, `Description`, `Group`, `Dir`, `Handoff bool`, `Actions []Action`
  - `Config` — `Title`, `Tasks []Task`
- **Sentinel**: `ErrNotFound` for missing config files
- **Task methods**: `AllCommands()`, `HasBackgroundActions()`, `HasHandoff()`, `LastAction()`
- **Functions**: `Load(path string)` reads YAML, `Write(path string, cfg Config)` saves YAML
- **Default title**: `"Nexus"`

#### Runner Package (`runner/` - ~1497 lines total)
Core command execution engine with three execution modes:

1. **Stream Mode** (default)
   - `Stream(task, lines chan LogLine)` — executes actions sequentially in foreground
   - Streams output to TUI in real-time via the `lines` channel
   - Blocks until all actions complete
   - Used for typical command workflows

2. **Background Mode**
   - `RunBackground(task)` — non-blocking, returns a `*BackgroundProc` immediately
   - Spawns all actions in a background process group
   - `BackgroundProc` struct tracks metadata and buffers output lines
   - Output accessible via the `BackgroundModel` TUI view

3. **Handoff Mode**
   - `HandoffLastAction(task)` — last action takes raw terminal control
   - Child process becomes interactive terminal session (Ctrl+C works natively)
   - Used for REPLs, editors, shells, and interactive tools

**Key Types:**
- `LogLine` — `Text string`, `IsErr bool` (stderr flag)
- `BackgroundProc` — metadata + output line buffer (1 MB scanner capacity per line)
- `ProcessRegistry` — thread-safe map of all spawned processes

**Key Components:**
- **`runner.go`** (431 lines): All three execution modes, `FormatCommand()`, `buildCmdFromShell()`
- **`registry.go`** (67 lines): Global `ProcessRegistry` singleton via `GetGlobalRegistry()`, `Register()`, `Unregister()`, `KillAll()`, `Count()`
- **`cmd_unix.go`** (58 lines): `sh -c` execution, `setpgid()` for process groups, SIGTERM → 100ms → SIGKILL cleanup
- **`cmd_windows.go`** (46 lines): `cmd.exe /C` execution, `CREATE_NEW_PROCESS_GROUP`, `taskkill /F /T /PID` for tree termination

#### UI Package (`ui/` - Bubble Tea TUI Framework)

State machine (`AppModel`) with three states managed in `app.go` (238 lines):

- **`stateMenu`** — fuzzy search list (default)
- **`stateOutput`** — real-time streaming output view
- **`stateBG`** — background process log viewer

**Key exported functions in `app.go`:**
- `NewApp(cfg *config.Config) AppModel` — constructs root model
- `Run(cfg *config.Config, cfgPath string) error` — starts the TUI with alt screen
- `RunFirstMatch(cfg *config.Config, query string) error` — fuzzy-matches and executes first hit, bypasses TUI entirely
- `HandleHandoff(task config.Task) error` — runs setup actions then hands off last action to raw terminal

**Keyboard shortcuts:**
- `q` or `Ctrl+C` in `stateMenu`: quit (kills all processes)
- `q` or `Ctrl+C` in `stateOutput`/`stateBG`: return to menu (kills all processes)
- `Esc` in `stateOutput`/`stateBG`: return to menu (kills all processes)

**`fuzzy.go`** (421 lines):
- `FuzzyModel` with real-time fuzzy filtering via `fuzzyMatch()`
- Dynamic grouping by `Task.Group` field
- Scroll/cursor/viewport management

**`output.go`** (314 lines):
- `OutputModel` — streaming log view with spinner animation, optional stderr highlighting
- `BackgroundModel` — background process log viewer with scroll

**`wizard.go`** (1109 lines):
- Two modes: CREATE (new config from scratch) and EDIT (modify existing config)
- **CREATE flow**: Welcome → project title → add command (name, description, commands, dir, group, run_mode) → loop for more → summary + confirm save
- **EDIT flow**: Hub screen with config summary, command list with delete checkboxes, action menu: `[d]` delete, `[e]` edit title, `[a]` add command, `[s]` save & exit
- 14 `wizStep` enum values
- Key types: `WizardModel`, `WizardResult`
- Key functions: `NewWizard()`, `NewWizardFromConfig()`, `RunWizard()`, `RunWizardEdit()`

**`list.go`** (92 lines) and **`group.go`** (145 lines): Legacy/alternate navigation components (not used in primary flow).

**`styles.go`** (59 lines): All shared Lipgloss style variables (colors, borders, layout, spacing).

**Test files** (8 files, ~4703 lines combined):
| File | Lines | Coverage |
|---|---|---|
| `app_test.go` | 685 | State transitions, handoff, RunFirstMatch |
| `fuzzy_test.go` | 455 | fuzzyMatch, filtering, navigation, scroll, viewport |
| `output_test.go` | 508 | OutputModel & BackgroundModel construction, update, scroll, rendering |
| `wizard_test.go` | 1544 | Full create/edit flows, delete step, hub actions, save/load round-trip |
| `list_test.go` | 265 | ListModel navigation, selection, view |
| `group_test.go` | 291 | GroupModel construction, navigation, selection, view |

### Configuration
- Commands are defined in YAML files (default: `nexus.yaml`)
- **Task structure**: Each task contains `name`, `description`, `actions[]` array, optional `dir`, `group`, and `handoff`
- **Actions format**: Each action has `command` string
- **Execution modes** (set at task level, not per-action):
  - **stream** (default): Runs actions sequentially, streams output in TUI, blocking
  - **background**: All actions run in background, returns immediately, output visible in background viewer
  - **handoff**: Last action takes over raw terminal control (Ctrl+C works as native terminal)
- Support for grouped commands (`group` field) and working directory override (`dir` field)
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
- Dark theme (`#0a0a0a` background) with fixed grid overlay effect
- Use Tailwind CSS v4 for styling; theme tokens defined via CSS custom properties (`oklch` color space) in `theme.css`
- Components in `web/src/app/components/` — one file per section
- Base path is `/nexus/` for GitHub Pages deployment
- Path alias `@` maps to `./src` (configured in `vite.config.ts`)
- `package.json` name is `@figma/my-make-file` (artifact from scaffolding — do not rely on it)

## CI/CD

### Automated Builds (`.github/workflows/build.yml` - 59 lines)
- **Triggers**: push to `main`, excluding `web/**` and `**.md` changes
- **Jobs**: cross-compile 5 binaries (Linux amd64/arm64, macOS amd64/arm64, Windows amd64), then create/update a GitHub release tagged `latest`

### Website Deployment (`.github/workflows/deploy.yml` - 56 lines)
- **Triggers**: push to `main`, only when `web/**` files change
- **Jobs**: checkout → Node 20 + `npm ci` → `npm run build` → deploy to GitHub Pages via `actions/deploy-pages`

## Testing

Always run quality checks before committing:
```bash
make check      # Run go vet and all tests
make lint       # Run golangci-lint
```

For comprehensive testing:
```bash
make test-cover         # Generate and open HTML coverage report
make test-cover-summary # Print per-package coverage summary to terminal
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

- `nexus.yaml`: Example configuration with inline documentation (119 lines)
- `Makefile`: All build and development commands (177 lines)
- `go.mod`: Go module definition (`module nexus`, Go 1.24.2)
- `web/vite.config.ts`: Vite configuration with `base: '/nexus/'` and `@` path alias
- `web/src/styles/theme.css`: Full CSS token theme using `oklch()` color space

## Cross-Platform Notes

The CLI supports:
- Windows (amd64) — `cmd.exe /C`, `CREATE_NEW_PROCESS_GROUP`, `taskkill /F /T`
- macOS (Intel amd64 and Apple Silicon arm64) — `sh -c`, `setpgid()`, SIGTERM/SIGKILL
- Linux (amd64 and arm64) — same as macOS

Platform-specific code:
- `runner/cmd_unix.go` — Unix/macOS/Linux
- `runner/cmd_windows.go` — Windows

Pre-built binaries are committed to `dist/` and also published as GitHub release assets tagged `latest`.

## License

MIT License (Copyright 2026 Sandro Šikić)
