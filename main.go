package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"os/signal"
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
	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	// Create server
	srv := server.New(logger, frontendFS)

	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server
	serverAddr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		serverAddr = ":" + port
	}

	if err := srv.Start(ctx, serverAddr); err != nil && err != context.Canceled {
		logger.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Application terminated")
}
