package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/cksidharthan/s3-browser/docs"
	"github.com/cksidharthan/s3-browser/internal/server"
)

//go:embed frontend/dist
var frontendFS embed.FS

// @title S3 Browser API
// @version 1.0
// @description A modern web-based file manager for S3-compatible storage systems
// @host localhost:8080
// @BasePath /api
func main() {
	// Define command-line flags
	var (
		port     = flag.String("port", "8080", "Port to run the server on")
		logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		help     = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	// Show help if requested
	if *help {
		fmt.Println("S3 Browser - A modern web-based file manager for S3-compatible storage")
		fmt.Println("")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  s3-browser")
		fmt.Println("  s3-browser -port 3000")
		fmt.Println("  s3-browser -port 8080 -log-level debug")
		fmt.Println("  s3-browser -help")
		os.Exit(0)
	}

	// Parse log level
	var logLevelVar slog.Level
	switch strings.ToLower(*logLevel) {
	case "debug":
		logLevelVar = slog.LevelDebug
	case "info":
		logLevelVar = slog.LevelInfo
	case "warn", "warning":
		logLevelVar = slog.LevelWarn
	case "error":
		logLevelVar = slog.LevelError
	default:
		fmt.Printf("Invalid log level: %s. Using 'info' instead.\n", *logLevel)
		logLevelVar = slog.LevelInfo
	}

	// Initialize structured logger with configurable level
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevelVar,
	}))

	// Log startup information
	logger.Info("Starting S3 Browser",
		slog.String("port", *port),
		slog.String("log_level", logLevelVar.String()),
	)

	// Create server
	srv := server.New(logger, frontendFS)

	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Determine server address
	serverAddr := ":" + *port
	// Environment variable takes precedence over flag for compatibility
	if envPort := os.Getenv("PORT"); envPort != "" {
		serverAddr = ":" + envPort
		logger.Info("Using PORT environment variable", slog.String("port", envPort))
	}

	if err := srv.Start(ctx, serverAddr); err != nil && err != context.Canceled {
		logger.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Application terminated")
}
