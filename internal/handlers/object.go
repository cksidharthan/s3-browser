package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cksidharthan/s3-browser/internal/middleware"
	"github.com/cksidharthan/s3-browser/internal/models"
)

// ObjectHandler handles object-related operations
type ObjectHandler struct {
	logger *slog.Logger
}

// NewObjectHandler creates a new object handler
func NewObjectHandler(logger *slog.Logger) *ObjectHandler {
	return &ObjectHandler{
		logger: logger,
	}
}

// ListObjects lists objects in a bucket
// @Summary List objects
// @Description Lists all objects in a specified S3 bucket
// @Tags Objects
// @Produce json
// @Param bucket query string true "Bucket name"
// @Success 200 {array} models.S3Object
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/objects [get]
func (h *ObjectHandler) ListObjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := middleware.GetSessionFromContext(ctx)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	result, err := session.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		h.logger.Error("Failed to list objects", 
			slog.String("bucket", bucket), 
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	objects := make([]models.S3Object, 0, len(result.Contents))
	for _, obj := range result.Contents {
		if obj.Key != nil {
			s3Object := models.S3Object{
				Key:  aws.ToString(obj.Key),
				Size: aws.ToInt64(obj.Size),
			}
			
			if obj.ETag != nil {
				s3Object.ETag = aws.ToString(obj.ETag)
			}
			if obj.StorageClass != "" {
				s3Object.StorageClass = string(obj.StorageClass)
			}
			if obj.LastModified != nil {
				s3Object.LastModified = obj.LastModified.Format("2006-01-02 15:04:05")
			}
			
			objects = append(objects, s3Object)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(objects)
}

// UploadObject uploads an object to S3
// @Summary Upload object
// @Description Uploads a file to the specified S3 bucket
// @Tags Objects
// @Accept multipart/form-data
// @Produce json
// @Param key path string true "Object key"
// @Param bucket query string true "Bucket name"
// @Param file formData file true "File to upload"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/objects/{key} [post]
func (h *ObjectHandler) UploadObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := middleware.GetSessionFromContext(ctx)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	key := h.extractObjectKeyFromPath(r.URL.Path)
	if key == "" {
		http.Error(w, "Object key is required", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Determine content type based on file extension
	contentType := "application/octet-stream"
	if header.Header.Get("Content-Type") != "" {
		contentType = header.Header.Get("Content-Type")
	} else {
		ext := strings.ToLower(filepath.Ext(header.Filename))
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".pdf":
			contentType = "application/pdf"
		case ".txt":
			contentType = "text/plain"
		case ".html", ".htm":
			contentType = "text/html"
		case ".css":
			contentType = "text/css"
		case ".js":
			contentType = "application/javascript"
		case ".json":
			contentType = "application/json"
		case ".xml":
			contentType = "application/xml"
		case ".zip":
			contentType = "application/zip"
		}
	}

	_, err = session.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: &header.Size,
	})
	if err != nil {
		h.logger.Error("Failed to upload object", 
			slog.String("bucket", bucket), 
			slog.String("key", key), 
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Object uploaded", 
		slog.String("bucket", bucket), 
		slog.String("key", key),
		slog.Int64("size", header.Size))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Object uploaded successfully",
		"bucket":  bucket,
		"key":     key,
	})
}

// ViewObject returns an object from S3
// @Summary View/Download object
// @Description Retrieves an object from S3 for viewing or download
// @Tags Objects
// @Param key path string true "Object key"
// @Param bucket query string true "Bucket name"
// @Success 200 "Object content"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Object not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/objects/{key} [get]
func (h *ObjectHandler) ViewObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := middleware.GetSessionFromContext(ctx)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	key := h.extractObjectKeyFromPath(r.URL.Path)
	if key == "" {
		http.Error(w, "Object key is required", http.StatusBadRequest)
		return
	}

	result, err := session.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		h.logger.Error("Failed to get object", 
			slog.String("bucket", bucket), 
			slog.String("key", key), 
			slog.String("error", err.Error()))
		
		// Determine status code based on error
		if strings.Contains(err.Error(), "NoSuchKey") {
			http.Error(w, "Object not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer result.Body.Close()

	// Set appropriate headers
	if result.ContentType != nil {
		w.Header().Set("Content-Type", aws.ToString(result.ContentType))
	}
	if result.ContentLength != nil {
		w.Header().Set("Content-Length", strconv.FormatInt(*result.ContentLength, 10))
	}
	
	// Set filename for download
	filename := filepath.Base(key)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

	// Copy object content to response
	_, err = io.Copy(w, result.Body)
	if err != nil {
		h.logger.Error("Failed to copy object content", 
			slog.String("bucket", bucket), 
			slog.String("key", key), 
			slog.String("error", err.Error()))
		return
	}

	h.logger.Info("Object served", 
		slog.String("bucket", bucket), 
		slog.String("key", key))
}

// DeleteObject deletes an object from S3
// @Summary Delete object
// @Description Deletes an object from the specified S3 bucket
// @Tags Objects
// @Param key path string true "Object key"
// @Param bucket query string true "Bucket name"
// @Success 204 "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Object not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/objects/{key} [delete]
func (h *ObjectHandler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := middleware.GetSessionFromContext(ctx)
	if session == nil {
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	key := h.extractObjectKeyFromPath(r.URL.Path)
	if key == "" {
		http.Error(w, "Object key is required", http.StatusBadRequest)
		return
	}

	_, err := session.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		h.logger.Error("Failed to delete object", 
			slog.String("bucket", bucket), 
			slog.String("key", key), 
			slog.String("error", err.Error()))
		
		// S3 delete doesn't fail if object doesn't exist, but handle other errors
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Object deleted", 
		slog.String("bucket", bucket), 
		slog.String("key", key))
	w.WriteHeader(http.StatusNoContent)
}

// GetPresignedURL generates a pre-signed URL for accessing an S3 object
// @Summary Get presigned URL
// @Description Generate a temporary URL for direct browser access to an S3 object
// @Tags Objects
// @Produce json
// @Param bucket query string true "Bucket name"
// @Param key query string true "Object key"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/presigned-url [get]
func (h *ObjectHandler) GetPresignedURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := middleware.GetSessionFromContext(ctx)
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
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
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
	json.NewEncoder(w).Encode(map[string]string{
		"url": presignResult.URL,
	})
}

// extractObjectKeyFromPath extracts object key from URL path like "/api/objects/{key}"
func (h *ObjectHandler) extractObjectKeyFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[len(parts)-2] == "objects" {
		return parts[len(parts)-1]
	}
	return ""
}
