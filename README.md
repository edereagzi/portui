# portui

Port-first process manager TUI for macOS, Linux, and Windows.

![portui demo](https://raw.githubusercontent.com/edereagzi/portui/main/docs/portui.gif)

## Installation

Install portui:

macOS / Linux (Recommended):

```bash
curl -fsSL https://raw.githubusercontent.com/edereagzi/portui/main/install.sh | sh
```

Windows (Recommended):

```powershell
irm https://raw.githubusercontent.com/edereagzi/portui/main/install.ps1 | iex
```

From source (Go 1.26+):

```bash
go install github.com/edereagzi/portui/cmd/portui@latest
```

Manual binaries (fallback):

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

```powershell
# Windows (amd64)
Invoke-WebRequest -Uri https://github.com/edereagzi/portui/releases/latest/download/portui_windows_amd64.zip -OutFile portui_windows_amd64.zip
Expand-Archive -Path .\portui_windows_amd64.zip -DestinationPath . -Force
New-Item -ItemType Directory -Force -Path "$HOME\bin" | Out-Null
Move-Item -Force .\portui.exe "$HOME\bin\portui.exe"
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

Self-update works when the current `portui` binary path is writable. On Windows, replacement is scheduled and completes after `portui` exits.

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

- Runtime: macOS/Linux (arm64 or amd64), Windows (amd64/arm64)
- Go is only required if you build from source (`go install` path)

## Windows notes

- On Windows, some listening sockets can appear with PID `0` or `4` due to Windows API ownership limitations.
- In those cases process name/user can be missing or mapped to system processes.
- This is a known Windows networking API limitation, not a TUI rendering bug.

## Why portui?

Every developer has typed `lsof -i :3000` to find what's eating their port. portui is the TUI that should have existed — see all listening ports mapped to their processes, navigate with vim keys, and kill with one keystroke. No more context switching, no more copy-pasting PIDs.

## Contributing

See `CONTRIBUTING.md` for branch naming and PR/issue linking workflow.

## License

MIT
