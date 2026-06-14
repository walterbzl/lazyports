# LazyPorts

> **Visual port manager for Linux, macOS, and Windows**

[![CI](https://github.com/walterbzl/lazyports/actions/workflows/ci.yml/badge.svg)](https://github.com/walterbzl/lazyports/actions/workflows/ci.yml)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)
[![GitHub stars](https://img.shields.io/github/stars/walterbzl/lazyports?style=flat-square)](https://github.com/walterbzl/lazyports/stargazers)
<br>
`lazyports` is a terminal UI tool to visualize and manage network ports. It provides an interactive table to inspect listening processes and kill them easily.

![LazyPorts Demo](assets/image.png)

Built with Bubble Tea and Lipgloss.

## Platform Support

LazyPorts selects a port scanner automatically at runtime based on your OS:

| Platform | Backend | Notes |
| :--- | :--- | :--- |
| **Linux** | `ss -tulnp` | Full support, including project auto-detection via `/proc`. |
| **macOS** | `lsof` | Full support. |
| **Windows** | `netstat -ano` + `tasklist` | Supported. CPU usage is not available from a single sample, so it shows `0%`. |

## Prerequisites

-   **Go 1.24** or higher (required for installation)

## Installation

### Linux / macOS — one-liner

```bash
curl -sL https://raw.githubusercontent.com/walterbzl/lazyports/main/install.sh | bash
```

*Note: the script may request `sudo` to install the binary to `/usr/local/bin` for system-wide access.*

### Any platform — `go install`

```bash
go install github.com/walterbzl/lazyports/cmd/lazyports@latest
```

The binary lands in `$(go env GOPATH)/bin` (add it to your `PATH` if needed).

### From source

```bash
git clone https://github.com/walterbzl/lazyports.git
cd lazyports
go build ./cmd/lazyports        # produces ./lazyports (or .\lazyports.exe on Windows)
# or run directly:
go run ./cmd/lazyports
```

On **Windows**, run `.\lazyports.exe` from PowerShell or Windows Terminal. If SmartScreen/Defender blocks the freshly built binary, use `go run ./cmd/lazyports` instead.

## Uninstallation

To remove `lazyports` from your system, simply delete the binary:

```bash
# Remove system-wide installation
sudo rm /usr/local/bin/lazyports

# Remove user-local installation
rm ~/go/bin/lazyports
```

## Usage

Run the tool from your terminal:

```bash
lazyports
```

### Controls

| Key | Action |
| :--- | :--- |
| `↑` / `↓` | Navigate list |
| `/` | **Filter / Search** by port, PID, process, or label |
| `Enter` | **View Details** (user, path, start time) |
| `Tab` | Toggle the **side detail panel** (connection, process, network, resources) |
| `k` | Kill selected process (graceful — SIGTERM / `taskkill`) |
| `K` | **Force kill** (SIGKILL / `taskkill /F`), with confirmation |
| `y` | **Copy** port number to clipboard |
| `o` | **Open** `http://localhost:<port>` in the browser |
| `l` | Assign a persistent **label** to the port |
| `r` | Refresh list |
| `R` | Toggle **auto-refresh** |
| `s` | **Sort** (cycle: Port → Process → PID) |
| `q` / `Ctrl+C` | Quit |

## Features

-   **Interactive Table**: Clean visualization of open ports (TCP/UDP).
-   **Side Detail Panel**: Press `Tab` for connection, process, network, and resource (CPU/MEM) info.
-   **Sortable Columns**: Press `s` to cycle sorting by Port, Process Name, or PID.
-   **Smart Filtering**: Type `/` to instantly filter by port, PID, process name, or label.
-   **Detailed Inspection**: Press `Enter` to see full command, user, and start time.

    ![Details View](assets/details.png)

-   **Process Management**: Graceful (`k`) and force (`K`) termination.
-   **Persistent Labels**: Tag ports with `l`; saved to `~/.config/lazyports/labels.json`.
-   **Quick Actions**: Copy a port (`y`) or open it in the browser (`o`).
-   **Auto-Refresh**: Toggle live updates with `R`, or set an interval in the config.
-   **Themes**: Catppuccin (default), Cherry Red, Tokyo Night, and Gruvbox.
-   **Visual States**: Distinct indicators for `LISTEN` and `ESTAB`, plus 🔒 local / 🌐 exposed.

## Configuration

LazyPorts works with zero config. To customize, copy the sample to your config directory:

```bash
mkdir -p ~/.config/lazyports
cp config.example.toml ~/.config/lazyports/config.toml
```

```toml
[general]
refresh_interval = 3        # seconds; 0 = manual refresh only
theme = "cherry"            # catppuccin | cherry | tokyo-night | gruvbox

[filters]
default_sort = "port"       # port | process | pid
```

## License

MIT
