package server

import (
	"context"
	"fmt"

	"litebase-backend/internal/ipc"
	"litebase-backend/internal/logger"
)

// Config holds the server configuration
type Config struct {
	SocketPath string
	PipeName   string
	Port       int
	Logger     logger.Logger
	DebugMode  bool // Enable debug mode for IPC server
}

// Server represents the main server
type Server struct {
	config *Config
	ipc    *ipc.Server
	logger logger.Logger
}

// New creates a new server instance
func New(config *Config) (*Server, error) {
	if config.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Create IPC server
	ipcConfig := &ipc.Config{
		SocketPath: config.SocketPath,
		PipeName:   config.PipeName,
		Logger:     config.Logger,
		DebugMode:  config.DebugMode,
	}

	ipcServer, err := ipc.New(ipcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPC server: %w", err)
	}

	server := &Server{
		config: config,
		ipc:    ipcServer,
		logger: config.Logger,
	}

	return server, nil
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Starting LiteBase Backend Server")

	// Start IPC server
	if err := s.ipc.Start(); err != nil {
		return fmt.Errorf("failed to start IPC server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	// Create a channel to signal when shutdown is complete
	done := make(chan error, 1)

	go func() {
		// Stop IPC server
		if err := s.ipc.Stop(); err != nil {
			done <- fmt.Errorf("failed to stop IPC server: %w", err)
			return
		}

		done <- nil
	}()

	// Wait for shutdown to complete or context to timeout
	select {
	case err := <-done:
		if err != nil {
			return err
		}
		s.logger.Info("Server shutdown completed successfully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("server shutdown timed out: %w", ctx.Err())
	}
}

// IsHealthy checks if the server is healthy
func (s *Server) IsHealthy() bool {
	// For now, just check if the server is running
	// In the future, this could check database connections, etc.
	return true
}
