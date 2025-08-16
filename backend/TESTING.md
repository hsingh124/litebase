# Testing the LiteBase Backend Locally

This guide shows you different ways to test the Go backend locally.

## 🚀 Quick Start (Recommended)

Use the automated test script:

```bash
# Make sure you're in the backend directory
cd backend

# Run the automated test
./test-local.sh
```

This script will:
- Build both the backend and test client
- Start the backend server
- Run the test client
- Clean up automatically when you press Ctrl+C

## 🔧 Manual Testing

### 1. Build the Backend

```bash
# Build the backend
go build -o build/litebase-backend ./main.go

# Build the test client
go build -o build/test-client ./cmd/test-client/main.go
```

### 2. Start the Backend Server

```bash
# Start with debug logging
./build/litebase-backend -log-level=debug

# Or start with specific socket path
./build/litebase-backend -socket=/tmp/my-socket.sock -log-level=debug
```

### 3. Test with the Test Client

In another terminal:

```bash
# Test with default Unix socket
./build/test-client

# Test with custom socket path
./build/test-client -socket=/tmp/my-socket.sock

# Test with TCP (if you want to use port instead)
./build/test-client -port=8080
```

## 🧪 Different Testing Approaches

### A. Unix Domain Socket Testing (Linux/macOS)

```bash
# Terminal 1: Start backend
./build/litebase-backend -log-level=debug

# Terminal 2: Test client
./build/test-client
```

### B. TCP Testing (Development)

```bash
# Terminal 1: Start backend with TCP port
./build/litebase-backend -port=8080 -log-level=debug

# Terminal 2: Test client with TCP
./build/test-client -port=8080
```

### C. Named Pipe Testing (Windows)

```bash
# Terminal 1: Start backend with named pipe
./build/litebase-backend -pipe=litebase-pipe -log-level=debug

# Terminal 2: Test client with named pipe
./build/test-client -pipe=litebase-pipe
```

## 🔍 What the Tests Do

The test client performs these tests:

1. **Health Check Test**
   - Sends a health check message
   - Expects a health response
   - Validates response format

2. **Unknown Message Test**
   - Sends an unknown message type
   - Expects an error response
   - Validates error handling

3. **Connection Management**
   - Tests connection establishment
   - Tests graceful disconnection
   - Tests message serialization/deserialization

## 📊 Expected Output

### Successful Test Run

```
🚀 LiteBase Backend Local Testing
==================================
[INFO] Go version: 1.21.0
[INFO] Building backend...
[SUCCESS] Backend built successfully!
[INFO] Building test client...
[SUCCESS] Test client built successfully!
[INFO] Starting backend server...
[SUCCESS] Backend started with PID: 12345
[SUCCESS] Unix socket created: /tmp/litebase.sock
[INFO] Testing backend with test client...
Connected to Unix socket: /tmp/litebase.sock
Testing health check...
✅ Health check passed!
Testing unknown message type...
✅ Unknown message test passed!
Keeping connection alive for 5 seconds...
Test completed successfully!
[SUCCESS] All tests passed! 🎉
```

### Backend Logs (in another terminal)

```json
{"level":"info","ts":1754860344.815785,"caller":"logger/logger.go:69","msg":"Starting LiteBase Backend","version":"1.0.0","buildTime":"unknown"}
{"level":"info","ts":1754860344.81637,"caller":"logger/logger.go:69","msg":"Starting LiteBase Backend Server"}
{"level":"info","ts":1754860344.816967,"caller":"logger/logger.go:69","msg":"IPC server started","address":"/tmp/litebase.sock"}
{"level":"debug","ts":1754860344.817123,"caller":"ipc/server.go:162","msg":"New connection established","remote":"@"}
{"level":"debug","ts":1754860344.817456,"caller":"ipc/server.go:230","msg":"Health check request received","id":"20240110121244.817123456"}
```

## 🐛 Troubleshooting

### Common Issues

1. **Permission Denied on Socket**
   ```bash
   # Check socket permissions
   ls -la /tmp/litebase.sock
   
   # Remove and recreate if needed
   rm /tmp/litebase.sock
   ./build/litebase-backend -log-level=debug
   ```

2. **Backend Won't Start**
   ```bash
   # Check if port is already in use
   lsof -i :8080
   
   # Use a different port
   ./build/litebase-backend -port=8081 -log-level=debug
   ```

3. **Test Client Can't Connect**
   ```bash
   # Check if backend is running
   ps aux | grep litebase-backend
   
   # Check socket file exists
   ls -la /tmp/litebase.sock
   
   # Restart backend
   pkill litebase-backend
   ./build/litebase-backend -log-level=debug
   ```

### Debug Mode

```bash
# Enable debug logging
./build/litebase-backend -log-level=debug

# Check logs for detailed information
tail -f /var/log/litebase-backend.log
```

## 🔄 Continuous Testing

For development, you can use the Go toolchain:

```bash
# Watch for changes and rebuild
go install github.com/air-verse/air@latest

# Create .air.toml configuration
air init

# Run with hot reload
air
```

## 📝 Adding New Tests

To add new test cases:

1. **Add new test function** in `cmd/test-client/main.go`
2. **Call the test** in the main function
3. **Rebuild** the test client
4. **Run** the tests

Example:

```go
func testDatabaseConnection(conn net.Conn) error {
    // Create database connection message
    msg := protocol.Message{
        ID:        time.Now().Format("20060102150405.000000000"),
        Type:      protocol.MessageTypeDBConnect,
        Timestamp: time.Now(),
        Data: map[string]interface{}{
            "driver": "postgres",
            "dsn":    "postgres://user:pass@localhost/db",
        },
    }
    
    // Send and validate...
    return nil
}
```

## 🎯 Next Steps

After successful local testing:

1. **Integration Testing** - Test with Tauri frontend
2. **Database Testing** - Test actual database connections
3. **Performance Testing** - Test with larger datasets
4. **Cross-Platform Testing** - Test on Windows and Linux
