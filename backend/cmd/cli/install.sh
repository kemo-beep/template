#!/bin/bash

# Mobile Backend CLI Installation Script
# This script installs the Mobile Backend CLI tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="mobile-backend-cli"
VERSION="1.0.0"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.mobile-backend-cli"
CONFIG_FILE="$CONFIG_DIR/config.yaml"

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check system requirements
check_requirements() {
    print_info "Checking system requirements..."
    
    # Check if Go is installed
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        print_info "Visit: https://golang.org/doc/install"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    REQUIRED_VERSION="1.21"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        print_error "Go version $GO_VERSION is not supported. Please install Go $REQUIRED_VERSION or later."
        exit 1
    fi
    
    # Check if Git is installed
    if ! command_exists git; then
        print_error "Git is not installed. Please install Git."
        print_info "Visit: https://git-scm.com/downloads"
        exit 1
    fi
    
    print_success "All requirements satisfied"
}

# Detect operating system
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS="linux";;
        Darwin*)    OS="darwin";;
        CYGWIN*)    OS="windows";;
        MINGW*)     OS="windows";;
        *)          OS="unknown";;
    esac
    
    case "$(uname -m)" in
        x86_64)     ARCH="amd64";;
        arm64)      ARCH="arm64";;
        aarch64)    ARCH="arm64";;
        *)          ARCH="unknown";;
    esac
    
    print_info "Detected OS: $OS, Architecture: $ARCH"
}

# Install from source
install_from_source() {
    print_info "Installing from source..."
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Clone repository (assuming this script is in the repo)
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    REPO_DIR="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"
    
    print_info "Using repository at: $REPO_DIR"
    
    # Build the CLI
    print_info "Building $BINARY_NAME..."
    cd "$REPO_DIR/backend/cmd/cli"
    
    # Install dependencies
    go mod tidy
    
    # Build
    go build -ldflags "-X main.version=$VERSION -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" -o "$BINARY_NAME" .
    
    # Install to system
    print_info "Installing to $INSTALL_DIR..."
    sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Cleanup
    cd /
    rm -rf "$TEMP_DIR"
    
    print_success "Installed $BINARY_NAME to $INSTALL_DIR"
}

# Install from binary release
install_from_binary() {
    print_info "Installing from binary release..."
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Download binary (this would be from a real release)
    BINARY_URL="https://github.com/your-org/mobile-backend-cli/releases/download/v$VERSION/$BINARY_NAME-$OS-$ARCH"
    
    print_info "Downloading binary from $BINARY_URL..."
    if command_exists curl; then
        curl -L -o "$BINARY_NAME" "$BINARY_URL"
    elif command_exists wget; then
        wget -O "$BINARY_NAME" "$BINARY_URL"
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    # Make executable
    chmod +x "$BINARY_NAME"
    
    # Install to system
    print_info "Installing to $INSTALL_DIR..."
    sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
    
    # Cleanup
    cd /
    rm -rf "$TEMP_DIR"
    
    print_success "Installed $BINARY_NAME to $INSTALL_DIR"
}

# Create configuration directory and files
setup_config() {
    print_info "Setting up configuration..."
    
    # Create config directory
    mkdir -p "$CONFIG_DIR"
    
    # Copy example config if it doesn't exist
    if [ ! -f "$CONFIG_FILE" ]; then
        SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        EXAMPLE_CONFIG="$SCRIPT_DIR/.mobile-backend-cli.yaml.example"
        
        if [ -f "$EXAMPLE_CONFIG" ]; then
            cp "$EXAMPLE_CONFIG" "$CONFIG_FILE"
            print_success "Created configuration file at $CONFIG_FILE"
        else
            # Create basic config
            cat > "$CONFIG_FILE" << EOF
# Mobile Backend CLI Configuration
base_url: "http://localhost:8080"
api_key: ""

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  name: "mobile_backend"
  driver: "postgres"

deployment:
  environments:
    development:
      region: "us-east-1"
      replicas: 1
    staging:
      region: "us-east-1"
      replicas: 2
    production:
      region: "us-west-2"
      replicas: 3
EOF
            print_success "Created basic configuration file at $CONFIG_FILE"
        fi
    else
        print_info "Configuration file already exists at $CONFIG_FILE"
    fi
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."
    
    if command_exists "$BINARY_NAME"; then
        VERSION_OUTPUT=$($BINARY_NAME --version 2>/dev/null || echo "unknown")
        print_success "Installation verified: $VERSION_OUTPUT"
        
        # Show help
        print_info "Showing help:"
        $BINARY_NAME --help
    else
        print_error "Installation verification failed. $BINARY_NAME not found in PATH."
        exit 1
    fi
}

# Show post-installation instructions
show_post_install() {
    print_success "Installation completed successfully!"
    echo
    print_info "Next steps:"
    echo "1. Configure your settings:"
    echo "   nano $CONFIG_FILE"
    echo
    echo "2. Test the installation:"
    echo "   $BINARY_NAME --version"
    echo "   $BINARY_NAME --help"
    echo
    echo "3. Start using the CLI:"
    echo "   $BINARY_NAME generate api User"
    echo "   $BINARY_NAME api explorer"
    echo "   $BINARY_NAME db migrate"
    echo
    print_info "For more information, visit: https://docs.example.com"
}

# Uninstall function
uninstall() {
    print_info "Uninstalling $BINARY_NAME..."
    
    # Remove binary
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        sudo rm -f "$INSTALL_DIR/$BINARY_NAME"
        print_success "Removed binary from $INSTALL_DIR"
    fi
    
    # Remove config (ask user)
    if [ -d "$CONFIG_DIR" ]; then
        read -p "Remove configuration directory $CONFIG_DIR? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$CONFIG_DIR"
            print_success "Removed configuration directory"
        else
            print_info "Configuration directory kept at $CONFIG_DIR"
        fi
    fi
    
    print_success "Uninstallation completed"
}

# Main installation function
main() {
    echo "🚀 Mobile Backend CLI Installer v$VERSION"
    echo "========================================"
    echo
    
    # Parse command line arguments
    case "${1:-}" in
        --uninstall)
            uninstall
            exit 0
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo
            echo "Options:"
            echo "  --uninstall    Uninstall the CLI"
            echo "  --help, -h     Show this help message"
            echo
            exit 0
            ;;
    esac
    
    # Check if already installed
    if command_exists "$BINARY_NAME"; then
        print_warning "$BINARY_NAME is already installed."
        read -p "Do you want to reinstall? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installation cancelled"
            exit 0
        fi
    fi
    
    # Run installation steps
    check_requirements
    detect_os
    
    # Choose installation method
    if [ "${2:-}" = "--binary" ]; then
        install_from_binary
    else
        install_from_source
    fi
    
    setup_config
    verify_installation
    show_post_install
}

# Run main function with all arguments
main "$@"
