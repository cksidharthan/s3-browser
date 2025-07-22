package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	_ "github.com/cksidharthan/s3-browser/docs"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

//go:generate cd frontend && npm run build && cd ..
//go:embed frontend/dist
var app embed.FS

// @title S3 API
// @version 1.0
// @description API for S3 operations
// @host localhost:8080
// @BasePath /
type S3Object struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ETag         string `json:"etag"`
	StorageClass string `json:"storage_class"`
}

type S3Bucket struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
}

// ConnectionRequest holds the parameters for establishing an S3 connection
type ConnectionRequest struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	UseSSL    bool   `json:"use_ssl"`
}

// ConnectionResponse represents the response from a connection attempt
type ConnectionResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// SessionStatusResponse represents the current session status
type SessionStatusResponse struct {
	HasSession bool `json:"has_session"`
}

// Session stores connection information
type Session struct {
	ID        string
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	UseSSL    bool
	S3Client  *s3.Client
	CreatedAt time.Time
	LastUsed  time.Time
}

// SessionManager manages user sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	logger   *slog.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(logger *slog.Logger) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		logger:   logger,
	}
}

// CreateSession creates a new session with the given connection parameters
func (sm *SessionManager) CreateSession(connReq ConnectionRequest) (*Session, error) {
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
	cfg, err := config.LoadDefaultConfig(context.Background(),
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

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	// Create session
	sessionID := uuid.New().String()
	session := &Session{
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
func (sm *SessionManager) GetSession(sessionID string) *Session {
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
func (sm *SessionManager) GetSessionFromCookie(r *http.Request) *Session {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil
	}
	return sm.GetSession(cookie.Value)
}

// DeleteSession removes a session by ID
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

// CleanupExpiredSessions removes sessions older than 24 hours
func (sm *SessionManager) CleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	expired := time.Now().Add(-24 * time.Hour)
	for id, session := range sm.sessions {
		if session.LastUsed.Before(expired) {
			delete(sm.sessions, id)
			sm.logger.Info("Session expired", slog.String("session_id", id))
		}
	}
}

// S3Handler handles requests for S3 operations
type S3Handler struct {
	sessionManager *SessionManager
	logger         *slog.Logger
}

// NewS3Handler initializes a new S3Handler with session management
func NewS3Handler(logger *slog.Logger) *S3Handler {
	sessionManager := NewSessionManager(logger)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			sessionManager.CleanupExpiredSessions()
		}
	}()

	return &S3Handler{sessionManager: sessionManager, logger: logger}
}

// requireSession middleware to check for valid session
func (h *S3Handler) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := h.sessionManager.GetSessionFromCookie(r)
		if session == nil {
			http.Error(w, "No valid session", http.StatusUnauthorized)
			return
		}
		// Store session in context for use by handlers
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// getSessionFromContext retrieves session from request context
func (h *S3Handler) getSessionFromContext(r *http.Request) *Session {
	if session, ok := r.Context().Value("session").(*Session); ok {
		return session
	}
	return nil
}

// getBucketName gets the bucket name from query parameters
func (h *S3Handler) getBucketName(r *http.Request) string {
	return r.URL.Query().Get("bucket")
}

// CheckSession checks if there's an active session
// @Summary Check session status
// @Description Checks if user has an active S3 session
// @Tags Session
// @Produce json
// @Success 200 {object} SessionStatusResponse
// @Router /api/session/status [get]
func (h *S3Handler) CheckSession(w http.ResponseWriter, r *http.Request) {
	session := h.sessionManager.GetSessionFromCookie(r)
	hasSession := session != nil

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SessionStatusResponse{HasSession: hasSession})
}

// Connect establishes an S3 connection and creates a session
// @Summary Connect to S3
// @Description Tests connection and creates a session if successful
// @Tags Session
// @Accept json
// @Produce json
// @Param connection body ConnectionRequest true "Connection parameters"
// @Success 200 {object} ConnectionResponse
// @Failure 400 {object} ConnectionResponse
// @Router /api/connect [post]
func (h *S3Handler) Connect(w http.ResponseWriter, r *http.Request) {
	var connReq ConnectionRequest

	// Parse request body
	err := json.NewDecoder(r.Body).Decode(&connReq)
	if err != nil {
		h.logger.Error("Failed to parse connection request", slog.String("error", err.Error()))
		sendConnectionResponse(w, false, "Invalid request format: "+err.Error(), "")
		return
	}

	// Validate required fields
	if connReq.Endpoint == "" || connReq.AccessKey == "" || connReq.SecretKey == "" || connReq.Region == "" {
		sendConnectionResponse(w, false, "Missing required fields: endpoint, access_key, secret_key, and region are required", "")
		return
	}

	h.logger.Info("Creating S3 connection",
		slog.String("endpoint", connReq.Endpoint),
		slog.String("region", connReq.Region),
		slog.String("access_key", connReq.AccessKey[:4]+"..."), // Only log first 4 chars for security
	)

	// Create session
	session, err := h.sessionManager.CreateSession(connReq)
	if err != nil {
		h.logger.Error("Failed to create session", slog.String("error", err.Error()))
		sendConnectionResponse(w, false, "Connection failed: "+err.Error(), "")
		return
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	h.logger.Info("Connection successful", slog.String("session_id", session.ID))
	sendConnectionResponse(w, true, "Connection successful", session.ID)
}

// Helper to send connection response
func sendConnectionResponse(w http.ResponseWriter, success bool, message string, sessionID string) {
	w.Header().Set("Content-Type", "application/json")

	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(ConnectionResponse{
		Success:   success,
		Message:   message,
		SessionID: sessionID,
	})
}

// Logout destroys the current session
// @Summary Logout
// @Description Destroys the current session and clears cookies
// @Tags Session
// @Success 200 {object} map[string]string
// @Router /api/logout [post]
func (h *S3Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session from cookie
	session := h.sessionManager.GetSessionFromCookie(r)
	if session != nil {
		// Delete session from memory
		h.sessionManager.DeleteSession(session.ID)
		h.logger.Info("Session destroyed", slog.String("session_id", session.ID))
	}

	// Clear the session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete the cookie
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// ListBuckets lists all available S3 buckets.
// @Summary List buckets
// @Description Lists all S3 buckets the user has access to
// @Tags S3
// @Accept json
// @Produce json
// @Success 200 {array} S3Bucket
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/buckets [get]
func (h *S3Handler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromContext(r)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	resp, err := session.S3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		h.logger.Error("Failed to list buckets", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buckets []S3Bucket
	for _, bucket := range resp.Buckets {
		buckets = append(buckets, S3Bucket{
			Name:         *bucket.Name,
			CreationDate: bucket.CreationDate.String(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buckets)
}

// CreateBucket creates a new S3 bucket.
// @Summary Create bucket
// @Description Creates a new S3 bucket with the given name
// @Tags S3
// @Accept json
// @Produce json
// @Param name path string true "Bucket Name"
// @Success 201 {string} string "Bucket created"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/buckets/{name} [put]
func (h *S3Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromContext(r)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	bucketName := vars["name"]

	_, err := session.S3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		h.logger.Error("Failed to create bucket", slog.String("bucket", bucketName), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Bucket created", slog.String("bucket", bucketName))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Bucket %s created successfully", bucketName)
}

// DeleteBucket deletes an S3 bucket.
// @Summary Delete bucket
// @Description Deletes an S3 bucket with the given name
// @Tags S3
// @Accept json
// @Produce json
// @Param name path string true "Bucket Name"
// @Success 204 "No Content"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/buckets/{name} [delete]
func (h *S3Handler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromContext(r)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	bucketName := vars["name"]

	_, err := session.S3Client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		h.logger.Error("Failed to delete bucket", slog.String("bucket", bucketName), slog.String("error", err.Error()))

		// Determine appropriate status code and user-friendly message based on error
		errorMessage := err.Error()
		statusCode := http.StatusInternalServerError
		userMessage := errorMessage

		// Check for common S3 errors and provide user-friendly messages
		if strings.Contains(errorMessage, "BucketNotEmpty") {
			statusCode = http.StatusConflict
			userMessage = "Cannot delete bucket: The bucket is not empty. Please delete all objects first."
		} else if strings.Contains(errorMessage, "NoSuchBucket") {
			statusCode = http.StatusNotFound
			userMessage = "Bucket not found."
		} else if strings.Contains(errorMessage, "AccessDenied") {
			statusCode = http.StatusForbidden
			userMessage = "Access denied: You don't have permission to delete this bucket."
		}

		http.Error(w, userMessage, statusCode)
		return
	}

	h.logger.Info("Bucket deleted", slog.String("bucket", bucketName))
	w.WriteHeader(http.StatusNoContent)
}

// ListObjects lists the objects in the specified S3 bucket.
// @Summary List objects
// @Description Retrieves a list of objects from the S3 bucket.
// @Tags S3
// @Accept json
// @Produce json
// @Param bucket query string true "Bucket name"
// @Success 200 {array} S3Object
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/objects [get]
func (h *S3Handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromContext(r)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := h.getBucketName(r)
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	resp, err := session.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		h.logger.Error("Failed to list objects", slog.String("bucket", bucket), slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize with empty slice to ensure JSON array response even when empty
	objects := make([]S3Object, 0)
	for _, obj := range resp.Contents {
		objects = append(objects, S3Object{
			Key:          *obj.Key,
			Size:         *obj.Size,
			ETag:         *obj.ETag,
			StorageClass: string(obj.StorageClass),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(objects)
}

// UploadObject uploads an object to the specified S3 bucket.
// @Summary Upload object
// @Description Uploads an object to the S3 bucket
// @Tags S3
// @Accept multipart/form-data
// @Produce json
// @Param bucket query string true "Bucket name"
// @Param key path string true "Object Key"
// @Param file formData file true "File to upload"
// @Success 201 {string} string "Object uploaded"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/objects/{key} [post]
func (h *S3Handler) UploadObject(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromContext(r)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := h.getBucketName(r)
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]

	// Parse the multipart form, 32 << 20 specifies a maximum upload of 32 MB
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		h.logger.Error("Failed to parse multipart form", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("Failed to get form file", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// If no key was provided in the URL, use the filename
	if key == "" {
		key = handler.Filename
	}

	// Upload the file to S3
	_, err = session.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:         aws.String(bucket),
		Key:            aws.String(key),
		Body:           file,
		ACL:            types.ObjectCannedACLPrivate,
		ContentType:    aws.String("application/json"),
		ChecksumSHA256: aws.String(""),
	})
	if err != nil {
		h.logger.Error("Failed to upload object",
			slog.String("bucket", bucket),
			slog.String("key", key),
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Object uploaded", slog.String("bucket", bucket), slog.String("key", key))
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "File %s uploaded successfully to bucket %s", key, bucket)
}

// DeleteObject deletes an object from the specified S3 bucket.
// @Summary Delete object
// @Description Deletes an object from the S3 bucket by key.
// @Tags S3
// @Accept json
// @Produce json
// @Param bucket query string true "Bucket name"
// @Param key path string true "Object Key"
// @Success 204 "No Content"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/objects/{key} [delete]
func (h *S3Handler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromContext(r)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := h.getBucketName(r)
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]

	_, err := session.S3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		h.logger.Error("Failed to delete object",
			slog.String("bucket", bucket),
			slog.String("key", key),
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Object deleted", slog.String("bucket", bucket), slog.String("key", key))
	w.WriteHeader(http.StatusNoContent)
}

// ViewObject retrieves an object from the specified S3 bucket.
// @Summary View object
// @Description Retrieves an object from the S3 bucket by key.
// @Tags S3
// @Accept json
// @Produce octet-stream
// @Param bucket query string true "Bucket name"
// @Param key path string true "Object Key"
// @Success 200 {string} string "Object data"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/objects/{key} [get]
func (h *S3Handler) ViewObject(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromContext(r)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := h.getBucketName(r)
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	key := vars["key"]

	getObjectResp, err := session.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		h.logger.Error("Failed to get object",
			slog.String("bucket", bucket),
			slog.String("key", key),
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer getObjectResp.Body.Close()

	// Set appropriate content type if available
	contentType := "application/octet-stream"
	if getObjectResp.ContentType != nil {
		contentType = *getObjectResp.ContentType
	} else {
		// Try to guess content type from file extension
		ext := filepath.Ext(key)
		if ext != "" {
			switch strings.ToLower(ext) {
			case ".html", ".htm":
				contentType = "text/html"
			case ".txt":
				contentType = "text/plain"
			case ".json":
				contentType = "application/json"
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".png":
				contentType = "image/png"
			case ".pdf":
				contentType = "application/pdf"
			}
		}
	}

	w.Header().Set("Content-Type", contentType)
	if getObjectResp.ContentLength != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *getObjectResp.ContentLength))
	}
	w.WriteHeader(http.StatusOK)
	io.Copy(w, getObjectResp.Body)
}

// GetPresignedURL generates a pre-signed URL for accessing an S3 object directly
// @Summary Get a pre-signed URL for an S3 object
// @Description Generate a temporary pre-signed URL that allows direct browser access to an S3 object
// @Tags S3
// @Accept json
// @Produce json
// @Param bucket query string true "S3 bucket name"
// @Param key query string true "Object key/path in the bucket"
// @Success 200 {object} map[string]string "Returns a JSON object with the pre-signed URL"
// @Failure 400 {string} string "Bad request - missing bucket or key parameter"
// @Failure 401 {string} string "Unauthorized - invalid or missing session"
// @Failure 500 {string} string "Internal server error"
// @Router /api/presigned-url [get]
func (h *S3Handler) GetPresignedURL(w http.ResponseWriter, r *http.Request) {
	session := h.getSessionFromContext(r)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Object key is required", http.StatusBadRequest)
		return
	}

	// Determine the content type based on file extension
	contentType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".pdf":
		contentType = "application/pdf"
	case ".mp4":
		contentType = "video/mp4"
	case ".mp3":
		contentType = "audio/mpeg"
	}

	// Create the presigner
	presignClient := s3.NewPresignClient(session.S3Client)

	// Generate a presigned URL with the content type explicitly set
	presignResult, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket:              aws.String(bucket),
		Key:                 aws.String(key),
		ResponseContentType: aws.String(contentType), // Force the correct content type
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})

	if err != nil {
		h.logger.Error("Failed to generate presigned URL",
			slog.String("bucket", bucket),
			slog.String("key", key),
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the presigned URL as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"url": presignResult.URL,
	})
}

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	// Create S3 handler with session management
	handler := NewS3Handler(logger)

	// Extract embedded frontend files
	dist, err := fs.Sub(app, "frontend/dist")
	if err != nil {
		logger.Error("Unable to read the frontend code", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Set up router
	router := mux.NewRouter()

	// API endpoints
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Session management endpoints (no auth required)
	apiRouter.HandleFunc("/session/status", handler.CheckSession).Methods("GET")
	apiRouter.HandleFunc("/connect", handler.Connect).Methods("POST")
	apiRouter.HandleFunc("/logout", handler.Logout).Methods("POST")

	// Protected S3 operations endpoints (require session)
	// Bucket operations
	apiRouter.HandleFunc("/buckets", handler.requireSession(handler.ListBuckets)).Methods("GET")
	apiRouter.HandleFunc("/buckets/{name}", handler.requireSession(handler.CreateBucket)).Methods("PUT")
	apiRouter.HandleFunc("/buckets/{name}", handler.requireSession(handler.DeleteBucket)).Methods("DELETE")

	// Object operations
	apiRouter.HandleFunc("/objects", handler.requireSession(handler.ListObjects)).Methods("GET")
	apiRouter.HandleFunc("/objects/{key}", handler.requireSession(handler.ViewObject)).Methods("GET")
	apiRouter.HandleFunc("/objects/{key}", handler.requireSession(handler.UploadObject)).Methods("POST")
	apiRouter.HandleFunc("/objects/{key}", handler.requireSession(handler.DeleteObject)).Methods("DELETE")
	apiRouter.HandleFunc("/presigned-url", handler.requireSession(handler.GetPresignedURL)).Methods("GET")

	// For backward compatibility (maintaining the original routes)
	apiRouter.HandleFunc("/list", handler.requireSession(handler.ListObjects)).Methods("GET")
	apiRouter.HandleFunc("/delete/{key}", handler.requireSession(handler.DeleteObject)).Methods("DELETE")
	apiRouter.HandleFunc("/view/{key}", handler.requireSession(handler.ViewObject)).Methods("GET")

	// Swagger documentation
	apiRouter.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Static frontend files - serve at root path
	router.PathPrefix("/").Handler(http.FileServer(http.FS(dist)))

	// Start server
	serverAddr := ":8080"
	logger.Info("Server starting", slog.String("address", serverAddr))
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		logger.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
