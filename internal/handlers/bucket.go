package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cksidharthan/s3-browser/internal/middleware"
	"github.com/cksidharthan/s3-browser/internal/models"
)

// BucketHandler handles bucket-related operations
type BucketHandler struct {
	logger *slog.Logger
}

// NewBucketHandler creates a new bucket handler
func NewBucketHandler(logger *slog.Logger) *BucketHandler {
	return &BucketHandler{
		logger: logger,
	}
}

// ListBuckets lists all buckets in the S3 account
// @Summary List buckets
// @Description Lists all S3 buckets accessible to the current session
// @Tags Buckets
// @Produce json
// @Success 200 {array} models.S3Bucket
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/buckets [get]
func (h *BucketHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := middleware.GetSessionFromContext(ctx)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	result, err := session.S3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		h.logger.Error("Failed to list buckets", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buckets := make([]models.S3Bucket, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		if bucket.Name != nil && bucket.CreationDate != nil {
			buckets = append(buckets, models.S3Bucket{
				Name:         aws.ToString(bucket.Name),
				CreationDate: bucket.CreationDate.Format("2006-01-02 15:04:05"),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buckets)
}

// CreateBucket creates a new S3 bucket
// @Summary Create bucket
// @Description Creates a new S3 bucket with the given name
// @Tags Buckets
// @Accept json
// @Produce json
// @Param name path string true "Bucket Name"
// @Success 201 "Created"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/buckets/{name} [put]
func (h *BucketHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := middleware.GetSessionFromContext(ctx)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucketName := h.extractBucketNameFromPath(r.URL.Path)
	if bucketName == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	_, err := session.S3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		h.logger.Error("Failed to create bucket", 
			slog.String("bucket", bucketName), 
			slog.String("error", err.Error()))
		
		// Determine appropriate status code based on error
		errorMessage := err.Error()
		statusCode := http.StatusInternalServerError
		userMessage := errorMessage
		
		if strings.Contains(errorMessage, "BucketAlreadyExists") {
			statusCode = http.StatusConflict
			userMessage = "Bucket already exists. Bucket names must be globally unique."
		} else if strings.Contains(errorMessage, "InvalidBucketName") {
			statusCode = http.StatusBadRequest
			userMessage = "Invalid bucket name. Bucket names must follow S3 naming conventions."
		} else if strings.Contains(errorMessage, "AccessDenied") {
			statusCode = http.StatusForbidden
			userMessage = "Access denied: You don't have permission to create buckets."
		}
		
		http.Error(w, userMessage, statusCode)
		return
	}

	h.logger.Info("Bucket created", slog.String("bucket", bucketName))
	w.WriteHeader(http.StatusCreated)
}

// DeleteBucket deletes an S3 bucket
// @Summary Delete bucket
// @Description Deletes an S3 bucket with the given name
// @Tags Buckets
// @Accept json
// @Produce json
// @Param name path string true "Bucket Name"
// @Success 204 "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 409 {string} string "Conflict - Bucket not empty"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/buckets/{name} [delete]
func (h *BucketHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := middleware.GetSessionFromContext(ctx)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucketName := h.extractBucketNameFromPath(r.URL.Path)
	if bucketName == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	_, err := session.S3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		h.logger.Error("Failed to delete bucket", 
			slog.String("bucket", bucketName), 
			slog.String("error", err.Error()))
		
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

// extractBucketNameFromPath extracts bucket name from URL path like "/api/buckets/{name}"
func (h *BucketHandler) extractBucketNameFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[len(parts)-2] == "buckets" {
		return parts[len(parts)-1]
	}
	return ""
}
