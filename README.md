# eDonish Auto — Go + Fyne

Automated grade management for [edonish.tj](https://edonish.tj).

## Features

- 🔐 Login with session persistence
- ⚙️ Auto-grade with configurable settings
- 📊 Journal viewer with grade display
- 📺 Real-time logs with copy support
- 🌗 Light/Dark theme toggle
- 🖥️ Cross-platform: Windows, macOS, Linux, Android

## Build from source

### Prerequisites

- Go 1.22+
- System dependencies (Linux):
  ```bash
  sudo apt install pkg-config libgl1-mesa-dev libxcursor-dev libxrandr-dev \
    libxinerama-dev libxi-dev libxxf86vm-dev libglu1-mesa-dev
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
├── api/       # eDonish API client (HTTP)
├── engine/    # Grade engine (concurrent workers)
└── ui/        # Fyne UI components
    ├── app.go       # Main app coordinator
    ├── login.go     # Login screen
    ├── auto.go      # Auto-grade page
    ├── journal.go   # Journal viewer
    └── logs.go      # Log viewer
```

## Memory

~30-50 MB RAM (vs 300+ MB with Flet/Python).

## License

Proprietary
