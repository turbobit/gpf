# gpf — Greenfield Port Forwarding

A fast, modern SSH port forwarding CLI & TUI tool built with Go and Bubble Tea.

![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue)
![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8)
![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

## Description

**gpf** (Greenfield Port Forwarding) is a terminal-native tool for managing SSH port forwards with ease. It supports both a minimal CLI and a full interactive TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea).

### TUI Preview

```
 ┌── gpf — Greenfield Port Forwarding ───────────────────────────┐
 │                                                                │
 │   Active Forwards                                            │
 │   ───────────────────────────────────────────────────────────  │
 │   ┌────────────┬──────────────────┬───────────┬───────┐       │
 │   │ Local      │ Remote           │ State     │       │       │
 │   ├────────────┼──────────────────┼───────────┼───────┤       │
 │   │ :8080      │ server.local:3000│ connected │       │       │
 │   │ :5432      │ db.local:5432    │ connected │       │       │
 │   │ :6379      │ cache.local:6379 │ stopping  │       │       │
 │   └────────────┴──────────────────┴───────────┴───────┘       │
 │                                                                │
 │   [New] [Connect] [Disconnect] [Quit]                         │
 └────────────────────────────────────────────────────────────────┘
```

## Features

- **Interactive TUI** — Visual dashboard for managing all port forwards at a glance
- **CLI mode** — Scriptable one-liners for CI/CD and automation
- **Multiple tunnels** — Forward multiple ports to different hosts simultaneously
- **Persistent sessions** — Survives connection drops with automatic reconnection
- **Cross-platform** — Works on Linux, macOS, and Windows
- **Zero dependencies** — Single statically-linked binary, no runtime deps
- **Fast startup** — Written in Go for near-instant launch times

## Installation

### Option 1: Go install

```bash
go install github.com/user/port-forwarding@latest
```

### Option 2: Unix install script

```bash
curl -sSfL https://raw.githubusercontent.com/user/port-forwarding/main/install/unix.sh | sh -s -- v0.1.0
```

Or with a specific version:

```bash
./install/unix.sh v0.1.0
```

### Option 3: Windows PowerShell

```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/user/port-forwarding/main/install/windows.ps1" -UseBasicParsing | Invoke-Expression
```

Or with a specific version:

```powershell
.\install\windows.ps1 v0.1.0
```

## Usage

### CLI Mode

#### Forward a single port

```bash
# Forward local:8080 to remote:3000
gpf forward localhost:8080 server:3000

# With custom SSH key
gpf forward --key ~/.ssh/id_ed25519 localhost:8080 server:3000

# With custom user
gpf forward --user admin localhost:8080 server:3000

# Bind to a specific interface
gpf forward --bind 0.0.0.0 localhost:8080 server:3000
```

#### Forward multiple ports

```bash
# Multiple local-to-remote mappings
gpf forward localhost:8080 web:3000 localhost:5432 db:5432 localhost:6379 cache:6379

# Or via config file
gpf forward --config forwards.yaml
```

#### Manage forwards

```bash
# List active forwards
gpf list

# Disconnect a specific forward
gpf disconnect 8080

# Disconnect all forwards
gpf disconnect --all
```

#### TUI Mode

```bash
# Launch the interactive terminal UI
gpf tui
```

### Commands

```
Usage:
  gpf [command]

Available Commands:
  forward     Create SSH port forwards
  disconnect  Disconnect an active forward
  list        List active port forwards
  tui         Launch the interactive TUI
  version     Print version info

Flags:
  -h, --help      Help for gpf
  -v, --version   Version for gpf
```

## Examples

### Development workflow

```bash
# Forward your app, database, and cache in one command
gpf forward \
  localhost:8080 app:3000 \
  localhost:5432 db:5432 \
  localhost:6379 redis:6379
```

### SSH config integration

```bash
# Use a specific SSH identity
gpf forward --key ~/.ssh/deploy_key \
  localhost:8443 staging:443
```

### One-off forward

```bash
# Connect, print the mapping, and exit
gpf forward --once localhost:9090 server:80
```

## Configuration

gpf looks for a configuration file in the following order:

1. `./gpf.yaml` (current directory)
2. `$HOME/.config/gpf/config.yaml`
3. `$HOME/.gpf/config.yaml`

Example config:

```yaml
ssh:
  user: deploy
  key: ~/.ssh/id_ed25519
  timeout: 10s

forwards:
  - local: "localhost:8080"
    remote: "production:3000"
  - local: "localhost:5432"
    remote: "production-db:5432"
```

## Building from source

```bash
git clone https://github.com/user/port-forwarding.git
cd port-forwarding
go build -o gpf .
```

## Internationalization (i18n)

gpf supports multiple languages through JSON translation files in the `i18n/` directory.

### Available Languages

| Language | File            |
|----------|-----------------|
| English  | `i18n/en.json`  |
| Korean   | `i18n/ko.json`  |

### Adding a New Language

1. Copy `i18n/en.json` to `i18n/<locale>.json` (e.g., `i18n/ja.json`, `i18n/fr.json`).
2. Translate all string values, leaving keys unchanged.
3. Save as UTF-8 encoded JSON.
4. Open a Pull Request with your new file.

### Contributing Translations

We welcome community translations! To contribute:

- Fork the repository.
- Add your `i18n/<locale>.json` file.
- Submit a PR with a brief description of the language and any notes.

See `i18n/README.md` for full details.

## License

MIT
