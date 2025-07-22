package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/cksidharthan/s3-browser/internal/models"
	"github.com/cksidharthan/s3-browser/internal/session"
)

// SessionHandler handles session-related operations
type SessionHandler struct {
	sessionManager *session.Manager
	logger         *slog.Logger
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionManager *session.Manager, logger *slog.Logger) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
		logger:         logger,
	}
}

// CheckSession checks if a valid session exists
// @Summary Check session status
// @Description Check if the current request has a valid session
// @Tags Session
// @Produce json
// @Success 200 {object} models.SessionStatusResponse
// @Router /api/session/status [get]
func (h *SessionHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
	session := h.sessionManager.GetSessionFromCookie(r)
	
	response := models.SessionStatusResponse{
		HasSession: session != nil,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Connect establishes a new S3 connection and creates a session
// @Summary Connect to S3
// @Description Establish connection to S3 storage and create session
// @Tags Session
// @Accept json
// @Produce json
// @Param connection body models.ConnectionRequest true "Connection parameters"
// @Success 200 {object} models.ConnectionResponse
// @Failure 400 {object} models.ConnectionResponse
// @Router /api/connect [post]
func (h *SessionHandler) Connect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var connReq models.ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&connReq); err != nil {
		h.sendConnectionResponse(w, false, "Invalid request format", "")
		return
	}

	// Validate required fields
	if connReq.Endpoint == "" || connReq.AccessKey == "" || connReq.SecretKey == "" || connReq.Region == "" {
		h.sendConnectionResponse(w, false, "Missing required connection parameters", "")
		return
	}

	h.logger.Info("Creating S3 connection",
		slog.String("endpoint", connReq.Endpoint),
		slog.String("region", connReq.Region),
		slog.String("access_key", connReq.AccessKey[:4]+"..."))

	// Create session with context
	session, err := h.sessionManager.CreateSession(ctx, connReq)
	if err != nil {
		h.logger.Error("Failed to create session", slog.String("error", err.Error()))
		h.sendConnectionResponse(w, false, "Failed to connect to S3: "+err.Error(), "")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true for HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	h.logger.Info("Connection successful", slog.String("session_id", session.ID))
	h.sendConnectionResponse(w, true, "Connection successful", session.ID)
}

// Logout destroys the current session
// @Summary Logout
// @Description Destroys the current session and clears cookies
// @Tags Session
// @Success 200 {object} map[string]string
// @Router /api/logout [post]
func (h *SessionHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session := h.sessionManager.GetSessionFromCookie(r)
	if session != nil {
		h.sessionManager.DeleteSession(session.ID)
		h.logger.Info("Session destroyed", slog.String("session_id", session.ID))
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// Helper to send connection response
func (h *SessionHandler) sendConnectionResponse(w http.ResponseWriter, success bool, message string, sessionID string) {
	w.Header().Set("Content-Type", "application/json")

	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(models.ConnectionResponse{
		Success:   success,
		Message:   message,
		SessionID: sessionID,
	})
}
