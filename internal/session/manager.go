package session

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"


	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cksidharthan/s3-browser/internal/models"
	"github.com/google/uuid"
)

// Manager manages user sessions
type Manager struct {
	sessions map[string]*models.Session
	mu       sync.RWMutex
	logger   *slog.Logger
}

// New creates a new session manager
func New(logger *slog.Logger) *Manager {
	return &Manager{
		sessions: make(map[string]*models.Session),
		logger:   logger,
	}
}

// CreateSession creates a new session with the given connection parameters
func (sm *Manager) CreateSession(ctx context.Context, connReq models.ConnectionRequest) (*models.Session, error) {
	// Fix endpoint URL if needed
	endpoint := connReq.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		if connReq.UseSSL {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}

	// Create AWS config
	credProvider := credentials.NewStaticCredentialsProvider(connReq.AccessKey, connReq.SecretKey, "")
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(connReq.Region),
		config.WithCredentialsProvider(credProvider),
		config.WithBaseEndpoint(endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.DisableLogOutputChecksumValidationSkipped = true
	})

	// Test the connection with timeout context
	testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err = client.ListBuckets(testCtx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	// Create session
	sessionID := uuid.New().String()
	session := &models.Session{
		ID:        sessionID,
		Endpoint:  endpoint,
		AccessKey: connReq.AccessKey,
		SecretKey: connReq.SecretKey,
		Region:    connReq.Region,
		UseSSL:    connReq.UseSSL,
		S3Client:  client,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	sm.logger.Info("Session created", slog.String("session_id", sessionID))
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *Manager) GetSession(sessionID string) *models.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil
	}

	// Update last used time
	session.LastUsed = time.Now()
	return session
}

// GetSessionFromCookie gets session from HTTP cookie
func (sm *Manager) GetSessionFromCookie(r *http.Request) *models.Session {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil
	}
	return sm.GetSession(cookie.Value)
}

// DeleteSession removes a session by ID
func (sm *Manager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

// CleanupExpiredSessions removes sessions older than 24 hours
func (sm *Manager) CleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	expiry := time.Now().Add(-24 * time.Hour)
	for id, session := range sm.sessions {
		if session.LastUsed.Before(expiry) {
			delete(sm.sessions, id)
			sm.logger.Info("Session expired and removed", slog.String("session_id", id))
		}
	}
}

// StartCleanupRoutine starts a background routine to clean up expired sessions
func (sm *Manager) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				sm.CleanupExpiredSessions()
			case <-ctx.Done():
				return
			}
		}
	}()
}
