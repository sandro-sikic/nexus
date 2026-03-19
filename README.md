# Nexus

**Your commands. One interface.**

Stop memorizing commands. Nexus transforms your scattered terminal commands into a clean, interactive TUI that works anywhere.

---

## Features

- **Interactive TUI** — Beautiful, auto-generated terminal interface from your YAML config. No manual setup required.
- **Lightning Fast** — Single binary with zero dependencies. Launch instantly and run commands without the overhead.
- **Simple Configuration** — Define commands in clean YAML. The setup wizard guides you through first-time configuration.
- **Cross-Platform** — Works seamlessly on Windows, macOS, and Linux. One tool for all your environments.
- **Open Source** — Fully transparent, community-driven development. Contribute, fork, or customize as you need.
- **Universal Commands** — Run any command: npm scripts, Docker commands, custom scripts, or system utilities.

---

## Getting Started

### 1. Download & Run

Download the binary for your platform from the [latest release](https://github.com/sandro-sikic/nexus/releases/latest). No installation, no dependencies.

| Platform | File |
|---|---|
| Linux (amd64) | `nexus_linux_amd64` |
| Linux (arm64) | `nexus_linux_arm64` |
| macOS (amd64) | `nexus_darwin_amd64` |
| macOS (arm64/Apple Silicon) | `nexus_darwin_arm64` |
| Windows (amd64) | `nexus_windows_amd64.exe` |

Make the binary executable on macOS/Linux:

```sh
chmod +x nexus_linux_amd64
./nexus_linux_amd64
```

### 2. Configure Commands

On first run, the setup wizard starts automatically when no configuration file is found:

```
$ nexus
⚠ No config found. Starting wizard...
```

You can also launch the wizard at any time:

```sh
nexus --wizard
```

The wizard walks you through:
- **Project Details** — Name your configuration and choose a default run mode
- **Commands** — Add commands with an optional group label and per-command run mode override

It generates a complete `nexus.yaml` for you — no manual YAML editing required.

### 3. Launch & Execute

Run Nexus anytime, select a command from the TUI menu, and execute with a single keystroke:

```
My Project Nexus

> █

── Development ──
▶ Dev Server        Start the development server with hot reload
    $ npm run dev
  Watch Tests       Run tests in watch mode

── Build ──
  Build             Production build
  Preview           Preview the production build locally

type to filter  •  ↑/↓ navigate  •  enter select  •  q quit
```

---

## Configuration

All commands are defined in a single `nexus.yaml` file:

```yaml
title: "My Project Nexus"
run_mode: stream      # default: stream output inside the TUI

commands:
  # ── Development ──────────────────────────────────────────
  - name: Dev Server
    description: Start the development server with hot reload
    command: "npm run dev"
    group: Development
    run_mode: handoff   # hand off terminal; Ctrl+C works normally

  - name: Setup and Dev
    description: Install deps then start the dev server
    commands:
      - "npm install"
      - "npm run dev"
    group: Development
    run_mode: handoff

  - name: Watch Tests
    description: Run tests in watch mode
    command: "npm run test:watch"
    group: Development
    run_mode: stream

  # ── Build ────────────────────────────────────────────────
  - name: Build
    description: Production build
    command: "npm run build"
    group: Build

  - name: Preview
    description: Preview the production build locally
    command: "npm run preview"
    group: Build
    run_mode: handoff

  # ── Docker ───────────────────────────────────────────────
  - name: Docker Up
    description: Start all Docker containers
    command: "docker compose up"
    group: Docker
    run_mode: handoff

  - name: Docker Logs
    description: View Docker container logs
    command: "docker compose logs -f"
    group: Docker
    run_mode: stream
```

---

## Building from Source

Requires Go 1.24+.

```sh
# Build for the current platform
make build

# Cross-compile for all platforms (output → dist/)
make build-all

# Run tests
make test

# Run vet + tests
make check
```

---

## License

MIT — see [LICENSE](LICENSE) for details.
