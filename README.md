# Go PowerControl

A modern, cross-platform MQTT power control application built with Go and Wails v2. This application manages power outlets on MQTT-connected power strips, providing real-time monitoring and control capabilities.

## ✨ Features

- 🌐 **Cross-Platform**: Runs on Windows and Linux
- 🔌 **MQTT Integration**: Connect to any MQTT broker to control power devices
- 📊 **Real-Time Monitoring**: Live status updates for all devices and outlets
- 🔍 **Smart Search**: Quick filtering across devices, outlets, and status
- 🔒 **Secure**: AES-256-GCM password encryption
- 🎨 **Modern UI**: Beautiful purple-themed interface built with web technologies
- 🔄 **Auto-Reconnect**: Automatic reconnection on network interruptions
- 📝 **Message Log**: Track all MQTT communications

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Node.js 16+** - [Download](https://nodejs.org/)
- **Wails CLI** - Install with: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **System Dependencies**:
  - **Windows**: [MinGW-w64](https://www.mingw-w64.org/) or [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
  - **Linux**: See [Linux Setup](#linux-setup) below for required packages

### Linux Setup

On Linux, Wails requires GTK3 and WebKit2GTK for building desktop applications. Follow these steps:

#### Quick Setup (Recommended)

Run the automated setup script:
```bash
./setup-linux.sh
```

This script will check and install all required dependencies, configure your PATH, and verify your setup.

#### Manual Setup

1. **Install system dependencies**:
   ```bash
   # Debian/Ubuntu/Linux Mint
   sudo apt install build-essential pkg-config libgtk-3-dev
   
   # Install WebKit2GTK (try 4.0 first, fallback to 4.1)
   sudo apt install libwebkit2gtk-4.0-dev || sudo apt install libwebkit2gtk-4.1-dev
   ```
   
   > [!NOTE]
   > Newer distributions (Debian 13+, Ubuntu 24.04+) only have `libwebkit2gtk-4.1-dev`. If you install 4.1, you'll need to create a compatibility symlink:
   > ```bash
   > sudo ln -sf webkit2gtk-4.1.pc /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
   > ```

2. **Install Wails CLI**:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

3. **Add Go binaries to your PATH**:
   
   The `wails` command will be installed to `$HOME/go/bin` (or `$GOPATH/bin`), which needs to be in your PATH.
   
   Add this line to your `~/.bashrc`, `~/.zshrc`, or equivalent shell config:
   ```bash
   export PATH="$HOME/go/bin:$PATH"
   ```
   
   Then reload your shell config:
   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```

4. **Verify installation**:
   ```bash
   wails doctor
   ```
   
   This should show all dependencies as "Installed" or "Available" (optional dependencies can be "Available").

### Installation from Source

1. **Clone the repository**
   ```bash
   git clone https://github.com/levonbragg/go-powercontrol.git
   cd go-powercontrol
   ```

2. **Install frontend dependencies**
   ```bash
   cd frontend
   npm install
   cd ..
   ```

3. **Run in development mode**
   ```bash
   wails dev
   ```

### Building for Production

#### Windows
```bash
wails build -platform windows/amd64
```
Or use the build script:
```bash
build\build-windows.bat
```

#### Linux
```bash
wails build -platform linux/amd64
```
Or use the build script:
```bash
chmod +x build/build-linux.sh
./build/build-linux.sh
```

Executables will be in `build/bin/`

## ⚙️ Configuration

On first run, the application will prompt you to configure your MQTT connection:

- **Username**: Your MQTT broker username
- **Password**: Your MQTT broker password (encrypted with AES-256-GCM)
- **MQTT Server**: Broker address (e.g., `192.168.1.100` or `mqtt.example.com`)
- **Port**: Broker port (default: 1883)
- **Subscribe String**: MQTT topic to subscribe to (default: `power/#`)

Configuration is stored in:
- **Windows**: `%APPDATA%\GoMQTTPowerControl\config.json`
- **Linux**: `~/.config/go-mqtt-power-control/config.json`

### Example Configuration

See `config.example.json` for a sample configuration file.

## 📡 MQTT Topic Structure

The application uses a well-defined topic hierarchy:

### Status Topics (received from devices)
```
power/<device-name>/outlets/<outlet-number>
```
**Example**: `power/office-strip/outlets/1`

### Command Topics (published to devices)
```
power/<device-name>/outlets/<outlet-number>/set
```
**Example**: `power/office-strip/outlets/1/set`

### Payload Values
- `0` = OFF
- `1` = ON

### Example Interaction

**Device publishes status**:
```
Topic: power/office-strip/outlets/1
Payload: 1
```

**User turns off the outlet**:
```
Topic: power/office-strip/outlets/1/set
Payload: 0
```

**Device confirms change**:
```
Topic: power/office-strip/outlets/1
Payload: 0
```

## 🎯 Usage

1. **Launch Application**: Start Go PowerControl
2. **Configure Connection**: Enter your MQTT broker details in the setup dialog
3. **Monitor Devices**: The grid will populate with devices as they publish status
4. **Search**: Use the search box to filter devices
5. **Control Outlets**: 
   - Click on a device/outlet row to select it
   - Choose desired state (ON/OFF) from dropdown
   - Click **Send** to publish command
6. **View Messages**: All MQTT communications are logged in the left panel

## 🏗️ Architecture

### Backend (Go)

- **`config/`**: Configuration management with AES-256 encryption
- **`mqtt/`**: MQTT client wrapper with auto-reconnect
- **`models/`**: Data structures for devices and messages
- **`app/`**: Wails application backend with bound methods

### Frontend (Svelte)

- **`frontend/src/App.svelte`**: Main application component
- **`frontend/src/components/`**: UI components
  - `MenuBar.svelte`: Application menu
  - `StatusBar.svelte`: Connection status indicator
  - `MessageLog.svelte`: MQTT message log
  - `DeviceGrid.svelte`: Device/outlet table with GroupByGrid behavior
  - `SearchBox.svelte`: Real-time search filter
  - `ControlPanel.svelte`: Outlet control interface
  - `SetupDialog.svelte`: Configuration settings
  - `AboutDialog.svelte`: Application info

## 🔒 Security

- **Password Encryption**: AES-256-GCM with machine-specific key
- **Secure Storage**: Config file with restricted permissions (0600)
- **No Plain Text**: Passwords are never stored unencrypted
- **Machine-Specific**: Encryption key derived from hostname and MAC address

## 🐛 Troubleshooting

### "Connection Failed"
- Verify MQTT broker is running and accessible
- Check server address and port
- Confirm username/password are correct
- Ensure firewall allows connection on the specified port

### "No Devices Found"
- Verify devices are publishing to the configured topic (`power/#`)
- Check topic structure matches expected format
- Confirm subscription was successful (check message log)

### "Failed to Save Config"
- Ensure application has permission to write to config directory
- Check disk space

### Build Issues

#### General
- Run `go mod tidy` to sync dependencies
- Run `npm install` in `frontend` directory

#### Linux-specific
- **"wails: command not found"**: 
  - The Wails CLI is installed but not in your PATH
  - Add `export PATH="$HOME/go/bin:$PATH"` to your shell config file (`~/.bashrc` or `~/.zshrc`)
  - Reload your shell: `source ~/.bashrc`
  - Verify: `which wails` should show the path to wails
  
- **Missing dependencies during build**:
  - Run `wails doctor` to check what's missing
  - Install required packages: `sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.0-dev`
  
- **"Package gtk+-3.0 was not found"**: Install GTK3 development headers:
  ```bash
  sudo apt install libgtk-3-dev
  ```

- **"Package webkit2gtk-4.0 was not found"**: 
  - On newer distributions, only `libwebkit2gtk-4.1-dev` is available
  - Install it: `sudo apt install libwebkit2gtk-4.1-dev`
  - Create a compatibility symlink:
    ```bash
    sudo ln -sf webkit2gtk-4.1.pc /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
    ```

#### Windows-specific
- Ensure MinGW-w64 or TDM-GCC is installed and in PATH
- Verify C compiler: `gcc --version`

## 📊 Performance

- **Startup Time**: < 2 seconds
- **MQTT Connection**: < 3 seconds (local network)
- **Message Throughput**: 100+ messages/second
- **Memory Usage**: < 100MB under normal operation
- **Search Filter**: < 100ms update time

## 🎨 Design Rationale

### Why Wails v2?
- Modern web-based UI with Go backend
- Cross-platform support (Windows, Linux, macOS)
- Smaller bundle size than Electron
- Native performance
- Familiar web development workflow

### Why Paho MQTT?
- Most popular Go MQTT client
- Excellent auto-reconnect support
- Well-maintained and documented
- Compatible with all MQTT brokers

### Why AES-256-GCM over TripleDES?
- Much more secure (modern standard)
- Better performance  
- No known vulnerabilities
- Future-proof

## 📝 Development

### Project Structure
```
go-powercontrol/
├── config/          # Configuration management
├── mqtt/            # MQTT client wrapper
├── models/          # Data structures
├── app/             # Wails backend
├── frontend/        # Svelte UI
├── build/           # Build scripts
├── assets/          # Application assets
└── main.go          # Entry point
```

### Adding New Features

1. **Backend**: Add methods to `app/app.go` and they'll be automatically bound to frontend
2. **Frontend**: Import bound methods from `../wailsjs/go/app/App`
3. **Events**: Use `runtime.EventsEmit()` to push updates to frontend

### Running Tests
```bash
go test ./...
```

## 📜 License

MIT License - see LICENSE file for details

## 👤 Author

Levon Bragg

## 🙏 Acknowledgments

- Original C# application: [Dataloggers-MQTT-MQTTnet](https://github.com/levonbragg/Dataloggers-MQTT-MQTTnet)
- Built with [Wails v2](https://wails.io/)
- MQTT client: [Paho MQTT Go](https://github.com/eclipse/paho.mqtt.golang)

---

**Note**: This is a complete rewrite of the original C# Windows Forms application. Config files are NOT backward compatible due to upgraded encryption (TripleDES+MD5 → AES-256-GCM).
