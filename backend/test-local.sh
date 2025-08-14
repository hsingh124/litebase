#!/bin/bash

# LiteBase Backend Local Testing Script
# This script helps you test the backend locally

set -e

echo "🚀 LiteBase Backend Local Testing"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_status "Go version: $GO_VERSION"

# Build the backend
print_status "Building backend..."
if go build -o build/litebase-backend ./main.go; then
    print_success "Backend built successfully!"
else
    print_error "Failed to build backend"
    exit 1
fi

# Build the test client
print_status "Building test client..."
if go build -o build/test-client ./cmd/test-client/main.go; then
    print_success "Test client built successfully!"
else
    print_error "Failed to build test client"
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    print_status "Cleaning up..."
    if [ ! -z "$BACKEND_PID" ]; then
        print_status "Stopping backend (PID: $BACKEND_PID)..."
        kill $BACKEND_PID 2>/dev/null || true
        wait $BACKEND_PID 2>/dev/null || true
    fi
    
    # Remove socket file if it exists
    if [ -S /tmp/litebase.sock ]; then
        print_status "Removing socket file..."
        rm -f /tmp/litebase.sock
    fi
    
    print_success "Cleanup completed!"
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

# Start the backend in background
print_status "Starting backend server..."
./build/litebase-backend -log-level=debug &
BACKEND_PID=$!

# Wait a moment for the backend to start
sleep 2

# Check if backend is running
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    print_error "Backend failed to start"
    exit 1
fi

print_success "Backend started with PID: $BACKEND_PID"

# Wait a bit more for socket creation
sleep 1

# Check if socket file was created
SOCKET_PATH=$(find /var/folders -name "litebase.sock" 2>/dev/null | head -1)
if [ -n "$SOCKET_PATH" ] && [ -S "$SOCKET_PATH" ]; then
	print_success "Unix socket created: $SOCKET_PATH"
	export SOCKET_PATH="$SOCKET_PATH"
else
	print_warning "Unix socket not found, checking if backend is running..."
	if ! kill -0 $BACKEND_PID 2>/dev/null; then
		print_error "Backend is not running"
		exit 1
	fi
fi

# Test the backend with our test client
print_status "Testing backend with test client..."
if [ -n "$SOCKET_PATH" ]; then
	if ./build/test-client -socket="$SOCKET_PATH"; then
		print_success "All tests passed! 🎉"
	else
		print_error "Some tests failed"
		exit 1
	fi
else
	print_error "Socket path not found"
	exit 1
fi

print_status "Test completed successfully!"
print_status "Backend is still running. Press Ctrl+C to stop it."
print_status "You can also run: ./build/test-client (in another terminal)"

# Keep the script running to keep the backend alive
wait $BACKEND_PID
