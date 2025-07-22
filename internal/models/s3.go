package models

// S3Object represents an S3 object with its metadata
type S3Object struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ETag         string `json:"etag"`
	StorageClass string `json:"storage_class"`
	LastModified string `json:"last_modified,omitempty"`
}

// S3Bucket represents an S3 bucket with its metadata
type S3Bucket struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
}
