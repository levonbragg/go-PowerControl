#!/bin/bash
# Quick setup script for Linux to install dependencies and configure PATH

set -e

echo "🔧 Go PowerControl - Linux Setup Script"
echo "========================================"
echo ""

# Check if running on Linux
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo "❌ This script is for Linux systems only"
    exit 1
fi

# Check for Go
echo "📦 Checking for Go..."
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first:"
    echo "   https://go.dev/dl/"
    exit 1
fi
echo "✅ Go $(go version | awk '{print $3}') found"

# Check for Node.js
echo "📦 Checking for Node.js..."
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 16+ first:"
    echo "   https://nodejs.org/"
    exit 1
fi
echo "✅ Node.js $(node --version) found"

# Check for required system dependencies
echo ""
echo "📦 Checking system dependencies..."
MISSING_DEPS=()

if ! dpkg -l | grep -q "build-essential"; then
    MISSING_DEPS+=("build-essential")
fi
if ! dpkg -l | grep -q "pkg-config"; then
    MISSING_DEPS+=("pkg-config")
fi
if ! dpkg -l | grep -q "libgtk-3-dev"; then
    MISSING_DEPS+=("libgtk-3-dev")
fi

# Check for webkit2gtk - try both 4.0 and 4.1
WEBKIT_INSTALLED=false
if dpkg -l | grep -q "libwebkit2gtk-4.0-dev"; then
    WEBKIT_INSTALLED=true
    WEBKIT_VERSION="4.0"
elif dpkg -l | grep -q "libwebkit2gtk-4.1-dev"; then
    WEBKIT_INSTALLED=true
    WEBKIT_VERSION="4.1"
fi

if ! $WEBKIT_INSTALLED; then
    # Try to determine which version is available
    if apt-cache show libwebkit2gtk-4.0-dev &> /dev/null; then
        MISSING_DEPS+=("libwebkit2gtk-4.0-dev")
    else
        MISSING_DEPS+=("libwebkit2gtk-4.1-dev")
    fi
fi

if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
    echo "⚠️  Missing system dependencies: ${MISSING_DEPS[*]}"
    echo ""
    read -p "Install missing dependencies? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "📥 Installing dependencies..."
        sudo apt update
        sudo apt install -y "${MISSING_DEPS[@]}"
        echo "✅ Dependencies installed"
        
        # Update webkit version if we just installed it
        if dpkg -l | grep -q "libwebkit2gtk-4.1-dev"; then
            WEBKIT_INSTALLED=true
            WEBKIT_VERSION="4.1"
        elif dpkg -l | grep -q "libwebkit2gtk-4.0-dev"; then
            WEBKIT_INSTALLED=true
            WEBKIT_VERSION="4.0"
        fi
    else
        echo "⚠️  Skipping dependency installation. Build may fail."
    fi
else
    echo "✅ All system dependencies are installed"
fi

# Handle webkit2gtk version compatibility
if $WEBKIT_INSTALLED && [ "$WEBKIT_VERSION" = "4.1" ]; then
    echo ""
    echo "📦 Checking WebKit2GTK compatibility..."
    if ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
        echo "⚠️  WebKit2GTK 4.1 detected, but Wails requires 4.0 reference"
        read -p "Create compatibility symlink? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            sudo ln -sf webkit2gtk-4.1.pc /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
            echo "✅ Compatibility symlink created"
        fi
    else
        echo "✅ WebKit2GTK compatibility OK"
    fi
fi

# Check for Wails CLI
echo ""
echo "📦 Checking for Wails CLI..."
GOBIN="${GOPATH:-$HOME/go}/bin"

if [ ! -f "$GOBIN/wails" ]; then
    echo "⚠️  Wails CLI not found"
    read -p "Install Wails CLI? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "📥 Installing Wails CLI..."
        go install github.com/wailsapp/wails/v2/cmd/wails@latest
        echo "✅ Wails CLI installed to $GOBIN/wails"
    else
        echo "⚠️  Skipping Wails installation. You'll need to install it manually."
    fi
else
    echo "✅ Wails CLI found at $GOBIN/wails"
fi

# Check if Go bin is in PATH
echo ""
echo "📦 Checking PATH configuration..."
if ! echo "$PATH" | grep -q "$GOBIN"; then
    echo "⚠️  $GOBIN is not in your PATH"
    echo ""
    echo "To use the 'wails' command, you need to add it to your PATH."
    echo ""
    
    # Detect shell
    SHELL_CONFIG=""
    if [ -n "$BASH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.bashrc"
    elif [ -n "$ZSH_VERSION" ]; then
        SHELL_CONFIG="$HOME/.zshrc"
    else
        echo "Unable to detect shell. Please manually add this line to your shell config:"
        echo "  export PATH=\"$GOBIN:\$PATH\""
        SHELL_CONFIG=""
    fi
    
    if [ -n "$SHELL_CONFIG" ]; then
        read -p "Add $GOBIN to PATH in $SHELL_CONFIG? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "" >> "$SHELL_CONFIG"
            echo "# Added by Go PowerControl setup script" >> "$SHELL_CONFIG"
            echo "export PATH=\"$GOBIN:\$PATH\"" >> "$SHELL_CONFIG"
            echo "✅ PATH updated in $SHELL_CONFIG"
            echo ""
            echo "⚠️  Please run: source $SHELL_CONFIG"
            echo "   Or restart your terminal for the changes to take effect"
        fi
    fi
else
    echo "✅ $GOBIN is in your PATH"
fi

# Final check with wails doctor
echo ""
echo "🏥 Running 'wails doctor' to verify setup..."
if command -v wails &> /dev/null; then
    wails doctor
else
    echo "⚠️  Cannot run 'wails doctor' yet. Please reload your shell or add $GOBIN to PATH"
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. cd frontend && npm install"
echo "  2. wails dev      # Run in development mode"
echo "  3. wails build    # Build for production"
