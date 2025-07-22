package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/cksidharthan/s3-browser/internal/models"
	"github.com/cksidharthan/s3-browser/internal/session"
)

// contextKey is used to store values in context
type contextKey string

const (
	SessionContextKey contextKey = "session"
)

// Auth provides authentication middleware
type Auth struct {
	sessionManager *session.Manager
	logger         *slog.Logger
}

// New creates a new auth middleware
func New(sessionManager *session.Manager, logger *slog.Logger) *Auth {
	return &Auth{
		sessionManager: sessionManager,
		logger:         logger,
	}
}

// RequireSession middleware to check for valid session
func (a *Auth) RequireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := a.sessionManager.GetSessionFromCookie(r)
		if session == nil {
			http.Error(w, "No valid session", http.StatusUnauthorized)
			return
		}

		// Store session in context for use by handlers
		ctx := context.WithValue(r.Context(), SessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetSessionFromContext retrieves session from request context
func GetSessionFromContext(ctx context.Context) *models.Session {
	if session, ok := ctx.Value(SessionContextKey).(*models.Session); ok {
		return session
	}
	return nil
}
