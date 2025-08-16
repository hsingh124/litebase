package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
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
		// Windows named pipe - this is a placeholder, actual implementation would use Windows-specific APIs
		log.Fatalf("Windows named pipe support not implemented yet")
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
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down client...")
		conn.Close()
		os.Exit(0)
	}()

	fmt.Println("\n🚀 Interactive LiteBase Backend Client")
	fmt.Println("=====================================")
	fmt.Println("Commands:")
	fmt.Println("  health     - Send health check")
	fmt.Println("  custom     - Send custom message")
	fmt.Println("  unknown    - Send unknown message type")
	fmt.Println("  quit       - Exit client")
	fmt.Println("  help       - Show this help")
	fmt.Println()

	// Start a goroutine to listen for server responses
	go listenForResponses(conn)

	// Interactive command loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("client> ")
		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}

		switch command {
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "help":
			showHelp()
		case "health":
			sendHealthCheck(conn)
		case "custom":
			sendCustomMessage(conn)
		case "unknown":
			sendUnknownMessage(conn)
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Type 'help' for available commands")
		}
	}
}

func showHelp() {
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  health     - Send health check message")
	fmt.Println("  custom     - Send custom message with data")
	fmt.Println("  unknown    - Send unknown message type (tests error handling)")
	fmt.Println("  quit/exit  - Exit the client")
	fmt.Println("  help       - Show this help message")
	fmt.Println()
}

func sendHealthCheck(conn net.Conn) {
	fmt.Println("📤 Sending health check...")

	msg := protocol.Message{
		ID:        time.Now().Format("20060102150405.000000000"),
		Type:      protocol.MessageTypeHealthCheck,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}

	if err := writeMessage(conn, &msg); err != nil {
		fmt.Printf("❌ Failed to send health check: %v\n", err)
		return
	}

	fmt.Println("✅ Health check sent successfully!")
}

func sendCustomMessage(conn net.Conn) {
	fmt.Println("📤 Sending custom message...")

	msg := protocol.Message{
		ID:        time.Now().Format("20060102150405.000000000"),
		Type:      "custom_message",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"action": "test",
			"value":  42,
			"items":  []string{"apple", "banana", "cherry"},
			"nested": map[string]interface{}{
				"key":    "value",
				"number": 123.45,
			},
		},
	}

	if err := writeMessage(conn, &msg); err != nil {
		fmt.Printf("❌ Failed to send custom message: %v\n", err)
		return
	}

	fmt.Println("✅ Custom message sent successfully!")
}

func sendUnknownMessage(conn net.Conn) {
	fmt.Println("📤 Sending unknown message type...")

	msg := protocol.Message{
		ID:        time.Now().Format("20060102150405.000000000"),
		Type:      "completely_unknown_type",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"reason": "testing error handling",
		},
	}

	if err := writeMessage(conn, &msg); err != nil {
		fmt.Printf("❌ Failed to send unknown message: %v\n", err)
		return
	}

	fmt.Println("✅ Unknown message sent successfully!")
}

func listenForResponses(conn net.Conn) {
	for {
		response, err := readMessage(conn)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\n🔌 Server disconnected")
				return
			}
			fmt.Printf("\n❌ Error reading response: %v\n", err)
			return
		}

		fmt.Printf("\n📥 Received response:\n")
		fmt.Printf("   ID: %s\n", response.ID)
		fmt.Printf("   Type: %s\n", response.Type)
		fmt.Printf("   Timestamp: %s\n", response.Timestamp.Format("15:04:05.000"))
		if len(response.Data) > 0 {
			fmt.Printf("   Data: %+v\n", response.Data)
		}
		fmt.Print("client> ")
	}
}

func writeMessage(conn net.Conn, msg *protocol.Message) error {
	// Serialize message
	data, err := msgpack.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write length prefix
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(data)))

	if _, err := conn.Write(lengthBytes); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}

	// Write message data
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	return nil
}

func readMessage(conn net.Conn) (*protocol.Message, error) {
	// Read length prefix (4 bytes)
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return nil, fmt.Errorf("EOF")
	}

	length := binary.BigEndian.Uint32(lengthBytes)

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}

	// Deserialize message
	var msg protocol.Message
	if err := msgpack.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}
