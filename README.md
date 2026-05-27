# gpf — Greenfield Port Forwarding

A fast, modern SSH port forwarding CLI & TUI tool built with Go and Bubble Tea.

![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-blue)
![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8)
![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)

**Documentation:** [English](README.md) | [한국어](README.ko.md)

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

### Option 1: GitHub Releases (Recommended)

Pre-built binaries are available on [GitHub Releases](https://github.com/turbobit/gpf/releases).

Download the appropriate binary for your platform:

| Platform | Binary |
|----------|--------|
| Linux amd64 | `gpf_linux_amd64` |
| Linux arm64 | `gpf_linux_arm64` |
| macOS arm64 | `gpf_darwin_arm64` |
| Windows amd64 | `gpf_windows_amd64.exe` |
| Windows arm64 | `gpf_windows_arm64.exe` |

```bash
# Example: Linux amd64
VERSION=v0.1.0
curl -LO "https://github.com/turbobit/gpf/releases/download/${VERSION}/gpf_linux_amd64"
chmod +x gpf_linux_amd64
sudo mv gpf_linux_amd64 /usr/local/bin/gpf
```

### Option 2: Go install

```bash
go install github.com/turbobit/gpf@latest
```

### Option 3: Unix install script

```bash
curl -sSfL https://raw.githubusercontent.com/turbobit/gpf/main/install/unix.sh | sh -s -- v0.1.0
```

Or install the latest version (no version argument needed):

```bash
curl -sSfL https://raw.githubusercontent.com/turbobit/gpf/main/install/unix.sh | sh
```

### Option 4: Windows PowerShell

```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/turbobit/gpf/main/install/windows.ps1" -UseBasicParsing | Invoke-Expression
```

Or with a specific version:

```powershell
.\install\windows.ps1 v0.1.0
```

## Usage

### Quick start

```bash
# Show all SSH servers from ~/.ssh/config
gpf

# Search servers by keyword (partial match on name, host, user)
gpf mac
gpf prod
gpf - macbook

# Use a specific language
gpf --lang ko
gpf -l en mac

# Scan listening ports on a server
gpf ports myserver

# Create a port forward
gpf forward myserver 3000        # remote :3000 → auto-assigned local port
gpf forward myserver 3000 8080   # remote :3000 → local :8080

# View active tunnels
gpf tunnels

# Stop a tunnel
gpf stop 12345                   # by PID
gpf stop-all
```

### Commands

| Command | Description |
|---------|-------------|
| `gpf` | Show all SSH servers (interactive TUI) |
| `gpf <keyword>` | Search servers (partial match, like `%keyword%`) |
| `gpf - <keyword>` | Same as above |
| `gpf ports <alias>` | Scan listening ports on a server |
| `gpf forward <alias> <remote-port> [local-port]` | Create a port forward |
| `gpf tunnels` | View and manage active tunnels |
| `gpf stop <pid>` | Stop a tunnel by PID |
| `gpf stop-all` | Stop all tunnels |
| `gpf version` | Show version info |
| `--lang <locale>` | Set UI language (`en`, `ko`) |
| `-l <locale>` | Short form of `--lang` |

### TUI keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate server list |
| `Enter` | Select action (Port Forward / SSH) |
| `f` | Forward selected port |
| `s` | SSH into selected server |
| `k` | Kill selected tunnel |
| `Ctrl+U` | Stop all tunnels |
| `r` | Refresh tunnel list |
| `/` | Filter servers |
| `Esc` | Go back |
| `q` | Quit (stops all tunnels) |

## Examples

### Quick port forward

```bash
# Forward production web server
gpf forward prod-web 3000

# Forward with a specific local port
gpf forward prod-db 5432 5432
```

### Search and connect

```bash
# Find all servers with "mac" in the name
gpf mac

# Find servers by host or user
gpf staging
gpf deploy
```

## Configuration

gpf reads your existing `~/.ssh/config` — no separate configuration file needed.

```
Host mac
  HostName 192.168.1.100
  User ubuntu
  Port 22
  IdentityFile ~/.ssh/id_ed25519

Host prod-web
  HostName web.example.com
  User deploy
```

## Building from source

```bash
git clone https://github.com/turbobit/gpf.git
cd gpf
go build -o gpf .
```

## Releases

gpf uses [GoReleaser](https://goreleaser.com/) with GitHub Actions to automate releases. When a `v*` tag is pushed, CI automatically builds binaries for Linux, macOS, and Windows and publishes them to GitHub Releases.

### Release workflow

1. A `v*` tag is pushed (e.g., `git tag v0.1.0 && git push origin v0.1.0`)
2. The **Release** GitHub Actions workflow triggers
3. GoReleaser cross-compiles for all supported platforms
4. Binaries are uploaded to a GitHub Release with generated changelog

### Supported platforms

| OS | Architectures |
|----|--------------|
| Linux | amd64, arm64 |
| macOS | arm64 |
| Windows | amd64, arm64 |

## Internationalization (i18n)

gpf supports multiple languages. The UI language is **automatically detected** from your system locale.

### How Language is Detected

gpf checks these environment variables in order:

1. `LANG` (e.g., `ko_KR.UTF-8` → Korean)
2. `LANGUAGE`
3. `LC_ALL`
4. `LC_MESSAGES`

If none of these are set, or the locale is not available, **English** is used as the default.

### Changing the Language

**Option 1: `--lang` flag (recommended)**

```bash
# Korean
gpf --lang ko
gpf -l ko mac          # Korean UI, search for "mac"
gpf tunnels --lang en  # English UI

# The flag works anywhere in the command
gpf forward prod 3000 --lang ko
```

**Option 2: Environment variable**

Set the `LANG` environment variable:

```bash
LANG=ko_KR.UTF-8 gpf
LANG=en_US.UTF-8 gpf

# Or permanently in your shell profile (~/.bashrc, ~/.zshrc)
export LANG=ko_KR.UTF-8
```

### Available Languages

| Language | File | Locale Example |
|----------|------|----------------|
| English  | `i18n/en.json` | `en`, `en_US`, `en_US.UTF-8` |
| Korean   | `i18n/ko.json` | `ko`, `ko_KR`, `ko_KR.UTF-8` |

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

## Acknowledgments

gpf was inspired by [ggh](https://github.com/byawitz/ggh), a wonderful SSH config helper. Thank you to [@byawitz](https://github.com/byawitz) for the inspiration.

## License

MIT
