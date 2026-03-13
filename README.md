# portui

Port-first process manager TUI for macOS and Linux — see what's on your ports and kill it with one keystroke.

![portui demo](docs/portui.gif)

## Installation

**Recommended (installer script, no Go required):**

```bash
curl -fsSL https://raw.githubusercontent.com/edereagzi/portui/main/install.sh | sh
```

Optional installer settings:

```bash
# Pin a specific version
curl -fsSL https://raw.githubusercontent.com/edereagzi/portui/main/install.sh | env VERSION=v0.1.1 sh

# Force install directory
curl -fsSL https://raw.githubusercontent.com/edereagzi/portui/main/install.sh | env INSTALL_DIR="$HOME/.local/bin" sh

# Disable checksum verification (not recommended)
curl -fsSL https://raw.githubusercontent.com/edereagzi/portui/main/install.sh | env VERIFY_CHECKSUM=0 sh
```

Installer environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `VERSION` | `latest` | Release tag to install (`v0.1.1`, `v0.2.0`, etc.) |
| `INSTALL_DIR` | auto (`/usr/local/bin` if writable, else `~/.local/bin`) | Destination directory |
| `VERIFY_CHECKSUM` | `1` | `1` verifies SHA256 against release `checksums.txt` |

**Manual binary install** (fallback):

```bash
# macOS (Apple Silicon)
curl -sL https://github.com/edereagzi/portui/releases/latest/download/portui_darwin_arm64.tar.gz | tar xz
sudo mv portui /usr/local/bin/

# macOS (Intel)
curl -sL https://github.com/edereagzi/portui/releases/latest/download/portui_darwin_amd64.tar.gz | tar xz
sudo mv portui /usr/local/bin/

# Linux (arm64)
curl -sL https://github.com/edereagzi/portui/releases/latest/download/portui_linux_arm64.tar.gz | tar xz
sudo mv portui /usr/local/bin/

# Linux (amd64)
curl -sL https://github.com/edereagzi/portui/releases/latest/download/portui_linux_amd64.tar.gz | tar xz
sudo mv portui /usr/local/bin/
```

**From source** (requires Go 1.26+):

```bash
go install github.com/edereagzi/portui/cmd/portui@latest
```

## Usage

Simply run:

```bash
portui
```

No flags or configuration needed. portui will scan your system for listening ports and display them in an interactive terminal UI.

Update to latest release:

```bash
portui update
```

Note: self-update supports direct installs on macOS/Linux where the current `portui` binary path is writable.

## Keybindings

| Key | Action |
|-----|--------|
| j / ↓ | Move down |
| k / ↑ | Move up |
| Enter | Toggle detail panel |
| x | Kill process |
| / | Search/filter |
| r | Refresh |
| ? | Help |
| q / Ctrl+C | Quit |

## Requirements

- macOS or Linux (arm64 or amd64)
- Go 1.26 or later (for building from source)

## Why portui?

Every developer has typed `lsof -i :3000` to find what's eating their port. portui is the TUI that should have existed — see all listening ports mapped to their processes, navigate with vim keys, and kill with one keystroke. No more context switching, no more copy-pasting PIDs.

## Contributing

See `CONTRIBUTING.md` for branch naming and PR/issue linking workflow.

## License

MIT
