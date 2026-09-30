package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for object storage (MinIO / S3 / Local).
type Storage interface {
	Upload(ctx context.Context, bucket, objectKey string, reader io.Reader, size int64, contentType string) (string, error)
	GetURL(ctx context.Context, bucket, objectKey string, expires time.Duration) (string, error)
	Delete(ctx context.Context, bucket, objectKey string) error
}

// Config holds object storage connection configurations.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}
