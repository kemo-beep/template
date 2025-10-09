#!/bin/bash

# Mobile Backend CLI Build Script

set -e

echo "🔨 Building Mobile Backend CLI..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Initialize Go module if it doesn't exist
if [ ! -f "go.mod" ]; then
    echo "📦 Initializing Go module..."
    go mod init mobile-backend-cli
fi

# Add required dependencies
echo "📦 Adding dependencies..."
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest

# Tidy dependencies
echo "🧹 Tidying dependencies..."
go mod tidy

# Build the CLI
echo "🔨 Building CLI binary..."
go build -ldflags "-X main.version=1.0.0 -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" -o mobile-backend-cli .

echo "✅ Build completed successfully!"
echo "📦 Binary created: mobile-backend-cli"
echo ""
echo "🚀 To install:"
echo "   sudo cp mobile-backend-cli /usr/local/bin/"
echo ""
echo "🧪 To test:"
echo "   ./mobile-backend-cli --help"
