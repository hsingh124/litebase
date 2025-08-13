package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"litebase-backend/internal/logger"
	"litebase-backend/internal/server"

	"go.uber.org/zap"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	// Parse command line flags
	var (
		socketPath = flag.String("socket", "", "Unix domain socket path (Linux/macOS)")
		pipeName   = flag.String("pipe", "", "Named pipe name (Windows)")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		port       = flag.Int("port", 0, "TCP port for development (optional)")
	)
	flag.Parse()

	// Initialize logger
	logger := logger.New(*logLevel)
	defer logger.Sync()

	logger.Info("Starting LiteBase Backend", zap.String("version", version), zap.String("buildTime", buildTime))

	// Create server configuration
	config := &server.Config{
		SocketPath: *socketPath,
		PipeName:   *pipeName,
		Port:       *port,
		Logger:     logger,
	}

	// Create and start server
	srv, err := server.New(config)
	if err != nil {
		logger.Error("Failed to create server", zap.Error(err))
		os.Exit(1)
	}

	// Start server in background
	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("Server failed to start", zap.Error(err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}
