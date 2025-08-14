package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"litebase-backend/internal/protocol"

	"github.com/vmihailenco/msgpack/v5"
)

func main() {
	var (
		socketPath = flag.String("socket", "/tmp/litebase.sock", "Unix domain socket path")
		pipeName   = flag.String("pipe", "", "Named pipe name (Windows)")
		port       = flag.Int("port", 0, "TCP port for development")
	)
	flag.Parse()

	// Determine connection method
	var conn net.Conn
	var err error

	if *port > 0 {
		// TCP connection for development
		conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", *port))
		if err != nil {
			log.Fatalf("Failed to connect to TCP server: %v", err)
		}
		fmt.Printf("Connected to TCP server on port %d\n", *port)
	} else if *pipeName != "" {
		// Windows named pipe
		conn, err = net.Dial("tcp", fmt.Sprintf("localhost:8080"))
		if err != nil {
			log.Fatalf("Failed to connect to named pipe: %v", err)
		}
		fmt.Printf("Connected to named pipe: %s\n", *pipeName)
	} else {
		// Unix domain socket
		conn, err = net.Dial("unix", *socketPath)
		if err != nil {
			log.Fatalf("Failed to connect to Unix socket: %v", err)
		}
		fmt.Printf("Connected to Unix socket: %s\n", *socketPath)
	}
	defer conn.Close()

	// Set up graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down client...")
		cancel()
		conn.Close()
		os.Exit(0)
	}()

	// Test health check
	fmt.Println("Testing health check...")
	if err := testHealthCheck(conn); err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("✅ Health check passed!")
	}

	// Test unknown message type
	fmt.Println("Testing unknown message type...")
	if err := testUnknownMessage(conn); err != nil {
		log.Printf("Unknown message test failed: %v", err)
	} else {
		fmt.Println("✅ Unknown message test passed!")
	}

	// Keep connection alive for a bit
	fmt.Println("Keeping connection alive for 5 seconds...")
	time.Sleep(5 * time.Second)
	fmt.Println("Test completed successfully!")
}

func testHealthCheck(conn net.Conn) error {
	// Create health check message
	msg := protocol.Message{
		ID:        time.Now().Format("20060102150405.000000000"),
		Type:      protocol.MessageTypeHealthCheck,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}

	// Send message
	if err := writeMessage(conn, &msg); err != nil {
		return fmt.Errorf("failed to write health check message: %v", err)
	}

	// Read response
	response, err := readMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to read health check response: %v", err)
	}

	// Validate response
	if response.Type != protocol.MessageTypeHealthResponse {
		return fmt.Errorf("unexpected response type: %s", response.Type)
	}

	fmt.Printf("Received response: %+v\n", response)
	return nil
}

func testUnknownMessage(conn net.Conn) error {
	// Create unknown message type
	msg := protocol.Message{
		ID:        time.Now().Format("20060102150405.000000000"),
		Type:      "unknown_type",
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}

	// Send message
	if err := writeMessage(conn, &msg); err != nil {
		return fmt.Errorf("failed to write unknown message: %v", err)
	}

	// Read response
	response, err := readMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to read error response: %v", err)
	}

	// Validate response
	if response.Type != protocol.MessageTypeError {
		return fmt.Errorf("expected error response, got: %s", response.Type)
	}

	fmt.Printf("Received error response: %+v\n", response)
	return nil
}

func writeMessage(conn net.Conn, msg *protocol.Message) error {
	// Serialize message
	data, err := msgpack.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Write length prefix
	length := uint32(len(data))
	lengthBytes := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	if _, err := conn.Write(lengthBytes); err != nil {
		return fmt.Errorf("failed to write message length: %v", err)
	}

	// Write message data
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %v", err)
	}

	return nil
}

func readMessage(conn net.Conn) (*protocol.Message, error) {
	// Read length prefix
	lengthBytes := make([]byte, 4)
	if _, err := conn.Read(lengthBytes); err != nil {
		return nil, fmt.Errorf("failed to read message length: %v", err)
	}

	length := uint32(lengthBytes[0])<<24 |
		uint32(lengthBytes[1])<<16 |
		uint32(lengthBytes[2])<<8 |
		uint32(lengthBytes[3])

	// Read message data
	data := make([]byte, length)
	if _, err := conn.Read(data); err != nil {
		return nil, fmt.Errorf("failed to read message data: %v", err)
	}

	// Deserialize message
	var msg protocol.Message
	if err := msgpack.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	return &msg, nil
}
