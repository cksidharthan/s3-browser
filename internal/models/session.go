package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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
