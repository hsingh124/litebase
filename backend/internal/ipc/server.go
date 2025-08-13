package ipc

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"litebase-backend/internal/logger"
	"litebase-backend/internal/protocol"

	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

// Server represents an IPC server
type Server struct {
	config   *Config
	listener net.Listener
	logger   logger.Logger
	handlers map[protocol.MessageType]MessageHandler
	ctx      context.Context
	cancel   context.CancelFunc
}

// Config holds the server configuration
type Config struct {
	SocketPath string
	PipeName   string
	Logger     logger.Logger
}

// MessageHandler is a function that handles incoming messages
type MessageHandler func(*protocol.Message) (*protocol.Message, error)

// New creates a new IPC server
func New(config *Config) (*Server, error) {
	if config.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		config:   config,
		logger:   config.Logger,
		handlers: make(map[protocol.MessageType]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Register default handlers
	server.registerDefaultHandlers()

	return server, nil
}

// Start starts the IPC server
func (s *Server) Start() error {
	var listener net.Listener
	var err error

	if runtime.GOOS == "windows" {
		// Windows: Use Named Pipes
		listener, err = s.createNamedPipeListener()
	} else {
		// Unix: Use Unix Domain Sockets
		listener, err = s.createUnixSocketListener()
	}

	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.listener = listener
	s.logger.Info("IPC server started", zap.String("address", listener.Addr().String()))

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				s.logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

// Stop stops the IPC server
func (s *Server) Stop() error {
	s.cancel()
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// createUnixSocketListener creates a Unix Domain Socket listener
func (s *Server) createUnixSocketListener() (net.Listener, error) {
	socketPath := s.config.SocketPath
	if socketPath == "" {
		// Default socket path
		socketPath = filepath.Join(os.TempDir(), "litebase.sock")
	}

	// Remove existing socket file if it exists
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Unix socket listener: %w", err)
	}

	// Set socket permissions
	if unixListener, ok := listener.(*net.UnixListener); ok {
		unixListener.SetUnlinkOnClose(true)
	}

	return listener, nil
}

// createNamedPipeListener creates a Named Pipe listener (Windows)
func (s *Server) createNamedPipeListener() (net.Listener, error) {
	pipeName := s.config.PipeName
	if pipeName == "" {
		pipeName = `\\.\pipe\litebase`
	}

	// For now, we'll use a TCP listener as a placeholder
	// In production, you'd want to use a proper Windows Named Pipe implementation
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %w", err)
	}

	s.logger.Warn("Using TCP listener instead of Named Pipes on Windows")
	return listener, nil
}

// handleConnection handles a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	s.logger.Debug("New connection established", zap.String("remote", conn.RemoteAddr().String()))

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// Read message
			msg, err := s.readMessage(conn)
			if err != nil {
				if err == io.EOF {
					s.logger.Debug("Connection closed by client")
					return
				}
				s.logger.Error("Failed to read message", zap.Error(err))
				return
			}

			// Handle message
			response, err := s.handleMessage(msg)
			if err != nil {
				s.logger.Error("Failed to handle message", zap.Error(err))
				errorResp := protocol.NewErrorResponse(err, 500, "Internal server error")
				response = &errorResp.Message
			}

			// Send response
			if err := s.writeMessage(conn, response); err != nil {
				s.logger.Error("Failed to write response", zap.Error(err))
				return
			}
		}
	}
}

// readMessage reads a MessagePack message from the connection
func (s *Server) readMessage(conn net.Conn) (*protocol.Message, error) {
	decoder := msgpack.NewDecoder(conn)
	var msg protocol.Message
	err := decoder.Decode(&msg)
	return &msg, err
}

// writeMessage writes a MessagePack message to the connection
func (s *Server) writeMessage(conn net.Conn, msg interface{}) error {
	encoder := msgpack.NewEncoder(conn)
	return encoder.Encode(msg)
}

// handleMessage routes messages to appropriate handlers
func (s *Server) handleMessage(msg *protocol.Message) (*protocol.Message, error) {
	handler, exists := s.handlers[msg.Type]
	if !exists {
		errorResp := protocol.NewErrorResponse(
			fmt.Errorf("unknown message type: %s", msg.Type),
			400,
			"Unsupported message type",
		)
		return &errorResp.Message, nil
	}

	return handler(msg)
}

// registerDefaultHandlers registers the default message handlers
func (s *Server) registerDefaultHandlers() {
	// Health check handler
	s.handlers[protocol.MessageTypeHealthCheck] = s.handleHealthCheck
}

// handleHealthCheck handles health check requests
func (s *Server) handleHealthCheck(msg *protocol.Message) (*protocol.Message, error) {
	s.logger.Debug("Health check request received", zap.String("id", msg.ID))

	response := protocol.NewHealthCheckResponse("healthy", "1.0.0")
	return &response.Message, nil
}
