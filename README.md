# portui

Port-first process manager TUI for macOS and Linux — see what's on your ports and kill it with one keystroke.

![portui demo](docs/portui.gif)

## Installation

**From binary** (no Go required):

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

**From source** (requires Go 1.22+):

```bash
go install github.com/edereagzi/portui/cmd/portui@latest
```

## Usage

Simply run:

```bash
portui
```

No flags or configuration needed. portui will scan your system for listening ports and display them in an interactive terminal UI.

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
- Go 1.22 or later (for building from source)

## Why portui?

Every developer has typed `lsof -i :3000` to find what's eating their port. portui is the TUI that should have existed — see all listening ports mapped to their processes, navigate with vim keys, and kill with one keystroke. No more context switching, no more copy-pasting PIDs.

## License

MIT
