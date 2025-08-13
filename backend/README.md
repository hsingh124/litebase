# LiteBase Backend

A high-performance Go backend for the LiteBase database client, designed to work as a Tauri sidecar application.

## Features

- **Cross-platform IPC**: Unix Domain Sockets (Linux/macOS) and Named Pipes (Windows)
- **MessagePack Protocol**: Fast binary serialization for IPC communication
- **Database Drivers**: Support for PostgreSQL, MySQL, and SQLite
- **Structured Logging**: Production-ready logging with zap
- **Graceful Shutdown**: Proper cleanup and resource management
- **Health Checks**: Built-in health monitoring endpoints

## Requirements

- Go 1.21 or higher
- Git (for version information)

## Quick Start

### 1. Build the Backend

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for specific platform
make build-linux
make build-macos
make build-windows
```

### 2. Run the Backend

```bash
# Development mode with debug logging
make dev

# Or run directly
go run ./main.go -log-level=debug

# Production mode
./build/litebase-backend -log-level=info
```

### 3. Command Line Options

```bash
Usage: litebase-backend [options]

Options:
  -socket string
        Unix domain socket path (Linux/macOS)
  -pipe string
        Named pipe name (Windows)
  -log-level string
        Log level (debug, info, warn, error) (default "info")
  -port int
        TCP port for development (optional)
```

## Architecture

### IPC Communication

The backend uses MessagePack for fast binary serialization over:

- **Unix Domain Sockets** on Linux/macOS
- **Named Pipes** on Windows (with TCP fallback for development)

### Message Protocol

All IPC messages follow this structure:

```go
type Message struct {
    ID        string                 `msgpack:"id"`
    Type      MessageType            `msgpack:"type"`
    Timestamp time.Time              `msgpack:"timestamp"`
    Data      map[string]interface{} `msgpack:"data"`
}
```

### Supported Message Types

- `health_check` - Health check request
- `health_response` - Health check response
- `db_connect` - Database connection request
- `db_connect_response` - Database connection response
- `query` - Query execution request
- `query_response` - Query execution response
- `error` - Error response

## Development

### Project Structure

```
backend/
├── main.go                 # Main entry point
├── Makefile               # Build automation
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── README.md              # This file
└── internal/              # Internal packages
    ├── ipc/              # IPC server implementation
    ├── logger/           # Structured logging
    ├── protocol/         # Message protocol definitions
    └── server/           # Main server coordination
```

### Development Commands

```bash
# Install development tools
make install-tools

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Update dependencies
make deps-update
```

### Adding New Message Types

1. Define the message type in `internal/protocol/message.go`
2. Add a handler function in `internal/ipc/server.go`
3. Register the handler in the `registerDefaultHandlers()` method

Example:

```go
// In protocol/message.go
const MessageTypeCustom MessageType = "custom"

type CustomMessage struct {
    Message
    CustomField string `msgpack:"custom_field"`
}

// In ipc/server.go
func (s *Server) handleCustom(msg *protocol.Message) (*protocol.Message, error) {
    // Handle custom message
    return response, nil
}

func (s *Server) registerDefaultHandlers() {
    s.handlers[protocol.MessageTypeCustom] = s.handleCustom
}
```

## Building for Production

### Cross-Platform Builds

```bash
# Build for all supported platforms
make build-all

# This creates binaries in the dist/ directory:
# - litebase-backend-linux-amd64
# - litebase-backend-linux-arm64
# - litebase-backend-darwin-amd64
# - litebase-backend-darwin-arm64
# - litebase-backend-windows-amd64.exe
# - litebase-backend-windows-arm64.exe
```

### Docker Build (Optional)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o litebase-backend ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/litebase-backend .
CMD ["./litebase-backend"]
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

```bash
# Test IPC communication
go test -v ./internal/ipc/...

# Test with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Performance

### Targets

- **Startup Time**: < 2 seconds
- **Memory Usage**: < 40MB idle, < 150MB with large datasets
- **IPC Latency**: < 1ms for small operations
- **Query Response**: < 100ms to start displaying results

### Monitoring

The backend includes structured logging for performance monitoring:

```bash
# Enable debug logging for performance analysis
./litebase-backend -log-level=debug
```

## Troubleshooting

### Common Issues

1. **Permission Denied on Unix Socket**
   ```bash
   # Check socket permissions
   ls -la /tmp/litebase.sock
   
   # Remove and recreate if needed
   rm /tmp/litebase.sock
   ```

2. **Port Already in Use (Windows)**
   ```bash
   # Use a different port
   ./litebase-backend -port=8081
   ```

3. **MessagePack Decoding Errors**
   - Ensure client and server use compatible MessagePack versions
   - Check message structure matches protocol definitions

### Debug Mode

```bash
# Enable debug logging
./litebase-backend -log-level=debug

# Check logs for detailed information
tail -f /var/log/litebase-backend.log
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Check the troubleshooting section
- Review the logs with debug level enabled
