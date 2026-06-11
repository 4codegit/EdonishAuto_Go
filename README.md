# eDonish Auto — Go + Fyne

Automated grade management for [edonish.tj](https://edonish.tj).

## Features

- 🔐 Login with session persistence
- 🏫 **Multi-school support** — teachers with multiple schools can switch between them
- ⚙️ Auto-grade with configurable settings
- 📊 Journal viewer with grade display (Н/А → "отсутствует")
- 📺 Real-time logs with copy support
- 🌗 Light/Dark theme toggle
- 🖥️ Cross-platform: Windows, macOS, Linux (.deb, .rpm), Android (.apk)

## Downloads

| Platform | File |
|----------|------|
| Fedora / RHEL | `.rpm` package |
| Ubuntu / Debian | `.deb` package |
| Linux (generic) | `edonish-auto` binary (amd64/arm64) |
| Windows | `edonish-auto.exe` |
| macOS (Intel) | `edonish-auto-mac` |
| macOS (Apple Silicon) | `edonish-auto-mac-arm64` |
| Android | `edonish-auto.apk` |

## Multi-School Support

If your account is linked to multiple schools, the app will:

1. Show a school selection screen after login
2. Display a school switcher in the dashboard header
3. Remember your last selected school on next login

## Build from source

### Prerequisites

- Go 1.22+
- System dependencies (Linux):
  ```bash
  # Debian/Ubuntu
  sudo apt install pkg-config libgl1-mesa-dev libxcursor-dev libxrandr-dev \
    libxinerama-dev libxi-dev libxxf86vm-dev libglu1-mesa-dev

  # Fedora
  sudo dnf install pkg-config mesa-libGL-devel libXcursor-devel libXrandr-devel \
    libXinerama-devel libXi-devel libXxf86vm-devel mesa-libGLU-devel
  ```

### Build

```bash
go build -o edonish-auto .
```

### Cross-compile

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o edonish-auto.exe .

# macOS
GOOS=darwin GOARCH=amd64 go build -o edonish-auto-mac .

# Linux ARM
GOOS=linux GOARCH=arm64 go build -o edonish-auto-arm64 .
```

### Android

```bash
fyne package -os android -appID com.edonish.auto
```

## Docker

```bash
docker build -t edonish-auto .
```

## Architecture

```
internal/
├── config/    # Configuration constants
├── api/       # eDonish API client (HTTP + multi-school)
├── engine/    # Grade engine (concurrent workers)
└── ui/        # Fyne UI components
    ├── app.go       # Main app coordinator
    ├── login.go     # Login screen
    ├── school.go    # School selection page
    ├── auto.go      # Auto-grade page
    ├── journal.go   # Journal viewer
    └── logs.go      # Log viewer
```

## Memory

~30-50 MB RAM (vs 300+ MB with Flet/Python).

## License

Proprietary
