# Unix Domain Sockets Explained

## 🔌 What's Happening in Our Backend

### Visual Representation

```
┌─────────────────┐    Socket File    ┌─────────────────┐
│   Go Backend    │◄─────────────────►│  Test Client    │
│   (Server)      │                   │   (Client)      │
└─────────────────┘                   └─────────────────┘
         │                                      │
         │                                      │
         ▼                                      ▼
┌─────────────────┐                   ┌─────────────────┐
│ Creates socket  │                   │ Connects to     │
│ file:           │                   │ socket file:    │
│ /tmp/litebase.  │                   │ /tmp/litebase.  │
│ sock            │                   │ sock            │
└─────────────────┘                   └─────────────────┘
```

## 🏗️ Step-by-Step Breakdown

### Step 1: Server Creates Socket
```go
// Backend starts up
func main() {
    // Create server configuration
    config := &server.Config{
        SocketPath: "/tmp/litebase.sock",  // ← This is where the socket file goes
        Logger:     logger,
    }
    
    // Start server
    srv.Start()  // ← This creates the socket file
}
```

**What happens:**
1. Backend creates a special "socket file" in the filesystem
2. This file acts like a "mailbox" for messages
3. Other processes can connect to this "mailbox"

### Step 2: Client Connects
```go
// Test client connects
conn, err := net.Dial("unix", "/tmp/litebase.sock")
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
```

**What happens:**
1. Client looks for the socket file
2. Establishes a connection through the socket
3. Can now send/receive messages

### Step 3: Message Exchange
```go
// Client sends a message
msg := protocol.Message{
    ID:   "123",
    Type: "health_check",
    Data: map[string]interface{}{},
}

// Message goes through the socket file
writeMessage(conn, &msg)

// Server receives and responds
response := handleHealthCheck(msg)
writeMessage(conn, response)
```

## 🔍 Why Use Socket Files Instead of Network Ports?

### ✅ **Advantages of Unix Sockets**

1. **Faster**: No network stack overhead
2. **More Secure**: Only local processes can connect
3. **File-based**: Easy to manage permissions
4. **No Port Conflicts**: No "port already in use" errors

### ❌ **Disadvantages**

1. **Unix-only**: Doesn't work on Windows (we use Named Pipes there)
2. **File System**: Depends on filesystem permissions
3. **Path Management**: Need to handle socket file paths

## 🌍 Cross-Platform Solution

Our backend handles both:

```go
func (s *Server) Start() error {
    var listener net.Listener
    var err error
    
    if runtime.GOOS == "windows" {
        // Windows: Use Named Pipes
        listener, err = s.createNamedPipeListener()
    } else {
        // Unix/Linux/macOS: Use Unix Domain Sockets
        listener, err = s.createUnixSocketListener()
    }
    
    // ... rest of the code
}
```

## 🔧 Socket File Details

### File Permissions
```bash
# Socket file permissions
ls -la /tmp/litebase.sock
# srwxr-xr-x  1 user  staff  0 Aug 13 23:53 litebase.sock
# ^^^^^^^^
# s = socket file
# rwx = owner can read/write/execute
# r-x = group can read/execute
# r-x = others can read/execute
```

### Socket File Properties
- **Type**: Special file (not regular file, not directory)
- **Size**: Usually 0 bytes (it's just a communication endpoint)
- **Permissions**: Control who can connect
- **Auto-cleanup**: Removed when server stops

## 🧪 Testing the Socket

### Check if Socket Exists
```bash
# Look for socket file
find /var/folders -name "litebase.sock" 2>/dev/null

# Check socket permissions
ls -la /var/folders/.../litebase.sock
```

### Test Connection
```bash
# Start backend
./build/litebase-backend -log-level=debug

# In another terminal, test connection
./build/test-client

# Or manually test with netcat (if available)
nc -U /var/folders/.../litebase.sock
```

## 🐛 Common Issues & Solutions

### 1. **Permission Denied**
```bash
# Check socket permissions
ls -la /tmp/litebase.sock

# Fix permissions if needed
chmod 666 /tmp/litebase.sock
```

### 2. **Socket File Not Found**
```bash
# Backend might be using a different path
# Check backend logs for actual socket path
./build/litebase-backend -log-level=debug

# Look for socket creation message:
# "IPC server started" address="/path/to/socket"
```

### 3. **Socket Already Exists**
```bash
# Remove old socket file
rm /tmp/litebase.sock

# Restart backend
./build/litebase-backend -log-level=debug
```

## 🔄 Alternative: TCP for Development

For easier debugging, you can use TCP instead:

```bash
# Start backend with TCP port
./build/litebase-backend -port=8080 -log-level=debug

# Test client connects via TCP
./build/test-client -port=8080
```

## 📚 Real-World Examples

### Database Connections
- **PostgreSQL**: Uses Unix sockets for local connections
- **MySQL**: Can use Unix sockets for local connections
- **Redis**: Uses Unix sockets for local connections

### System Services
- **Docker**: Uses Unix sockets for daemon communication
- **systemd**: Uses Unix sockets for service communication
- **X11**: Uses Unix sockets for display communication

## 🎯 Why This Matters for LiteBase

1. **Performance**: Fast local communication between Tauri frontend and Go backend
2. **Security**: No network exposure, only local processes can connect
3. **Reliability**: No network issues, port conflicts, or firewall problems
4. **Cross-platform**: Works on all major operating systems

## 🔍 Debugging Socket Issues

### Enable Debug Logging
```bash
./build/litebase-backend -log-level=debug
```

### Check Socket State
```bash
# See all Unix sockets
lsof -U

# See specific socket
lsof /tmp/litebase.sock
```

### Monitor Socket Traffic
```bash
# Watch socket file creation
watch -n 1 'ls -la /tmp/litebase.sock 2>/dev/null || echo "Socket not found"'
```

This socket-based architecture gives us a robust, fast, and secure way for the Tauri frontend to communicate with the Go backend for database operations! 🚀
