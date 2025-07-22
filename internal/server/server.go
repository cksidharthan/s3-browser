package server

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/cksidharthan/s3-browser/internal/handlers"
	"github.com/cksidharthan/s3-browser/internal/middleware"
	"github.com/cksidharthan/s3-browser/internal/session"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server represents the HTTP server
type Server struct {
	sessionManager *session.Manager
	auth           *middleware.Auth
	sessionHandler *handlers.SessionHandler
	bucketHandler  *handlers.BucketHandler
	objectHandler  *handlers.ObjectHandler
	logger         *slog.Logger
	mux            *http.ServeMux
}

// New creates a new server instance
func New(logger *slog.Logger, frontendFS embed.FS) *Server {
	sessionManager := session.New(logger)
	auth := middleware.New(sessionManager, logger)

	server := &Server{
		sessionManager: sessionManager,
		auth:           auth,
		sessionHandler: handlers.NewSessionHandler(sessionManager, logger),
		bucketHandler:  handlers.NewBucketHandler(logger),
		objectHandler:  handlers.NewObjectHandler(logger),
		logger:         logger,
		mux:            http.NewServeMux(),
	}

	server.setupRoutes(frontendFS)
	return server
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes(frontendFS embed.FS) {
	// Session management endpoints (no auth required)
	s.mux.HandleFunc("/api/session/status", s.sessionHandler.CheckSession)
	s.mux.HandleFunc("/api/connect", s.requireMethod(s.sessionHandler.Connect, http.MethodPost))
	s.mux.HandleFunc("/api/logout", s.requireMethod(s.sessionHandler.Logout, http.MethodPost))

	// Protected bucket endpoints
	s.mux.HandleFunc("/api/buckets", s.requireMethod(s.auth.RequireSession(s.bucketHandler.ListBuckets), http.MethodGet))
	s.mux.HandleFunc("/api/buckets/", s.handleBucketOperations)

	// Protected object endpoints
	s.mux.HandleFunc("/api/objects", s.requireMethod(s.auth.RequireSession(s.objectHandler.ListObjects), http.MethodGet))
	s.mux.HandleFunc("/api/objects/", s.handleObjectOperations)
	s.mux.HandleFunc("/api/presigned-url", s.requireMethod(s.auth.RequireSession(s.objectHandler.GetPresignedURL), http.MethodGet))

	// Swagger documentation
	s.mux.Handle("/api/swagger/", httpSwagger.WrapHandler)

	// Frontend static files
	s.setupFrontendRoutes(frontendFS)
}

// handleBucketOperations handles bucket operations based on HTTP method
func (s *Server) handleBucketOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		s.auth.RequireSession(s.bucketHandler.CreateBucket)(w, r)
	case http.MethodDelete:
		s.auth.RequireSession(s.bucketHandler.DeleteBucket)(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleObjectOperations handles object operations based on HTTP method
func (s *Server) handleObjectOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.auth.RequireSession(s.objectHandler.ViewObject)(w, r)
	case http.MethodPost:
		s.auth.RequireSession(s.objectHandler.UploadObject)(w, r)
	case http.MethodDelete:
		s.auth.RequireSession(s.objectHandler.DeleteObject)(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}



// setupFrontendRoutes configures static file serving for the frontend
func (s *Server) setupFrontendRoutes(frontendFS embed.FS) {
	// Extract embedded frontend files
	dist, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		s.logger.Error("Unable to read the frontend code", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Serve static files at root path
	fileServer := http.FileServer(http.FS(dist))
	s.mux.Handle("/", fileServer)
}

// requireMethod ensures only specified HTTP methods are allowed
func (s *Server) requireMethod(handler http.HandlerFunc, allowedMethods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, method := range allowedMethods {
			if r.Method == method {
				handler(w, r)
				return
			}
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context, addr string) error {
	// Start session cleanup routine
	s.sessionManager.StartCleanupRoutine(ctx)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Server starting", slog.String("address", addr))

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Server shutdown requested")

		// Graceful shutdown with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Server shutdown error", slog.String("error", err.Error()))
			return err
		}

		s.logger.Info("Server stopped gracefully")
		return ctx.Err()
	case err := <-errCh:
		s.logger.Error("Server error", slog.String("error", err.Error()))
		return err
	}
}
