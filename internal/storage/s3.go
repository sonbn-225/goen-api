package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Config holds SeaweedFS / S3-compatible configuration.
type S3Config struct {
	Endpoint      string // e.g. sunflower-seaweedfs:8333
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseSSL        bool
	PublicBaseURL string // e.g. https://api.example.com (for proxied URLs)
}

// S3Client wraps a minio client for SeaweedFS.
type S3Client struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

// NewS3Client creates a new S3Client. Returns nil if endpoint is empty (graceful degradation).
func NewS3Client(cfg S3Config) *S3Client {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil
	}
	bucket := cfg.Bucket
	if bucket == "" {
		bucket = "goen"
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil
	}

	s := &S3Client{client: client, bucket: bucket, publicBaseURL: cfg.PublicBaseURL}

	// Ensure bucket exists (non-fatal: log but don't crash).
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if exists, err := client.BucketExists(ctx, bucket); err == nil && !exists {
		_ = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}

	return s
}

// UploadAvatar uploads a multipart file header to SeaweedFS under avatars/<userID>/<uuid>.<ext>.
// Returns the object path (key) which can be used to build the public URL.
func (s *S3Client) UploadAvatar(ctx context.Context, userID string, file *multipart.FileHeader) (string, error) {
	if s == nil {
		return "", fmt.Errorf("storage not configured")
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	objectKey := fmt.Sprintf("avatars/%s/%s%s", userID, uuid.NewString(), ext)

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.client.PutObject(ctx, s.bucket, objectKey, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return objectKey, nil
}

// S3ObjectInfo contains metadata about a retrieved object.
type S3ObjectInfo struct {
	Size        int64
	ContentType string
}

// GetObject retrieves an object from SeaweedFS. Caller must close the returned ReadCloser.
func (s *S3Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, S3ObjectInfo, error) {
	if s == nil {
		return nil, S3ObjectInfo{}, fmt.Errorf("storage not configured")
	}
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, S3ObjectInfo{}, err
	}
	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, S3ObjectInfo{}, err
	}
	return obj, S3ObjectInfo{Size: stat.Size, ContentType: stat.ContentType}, nil
}

// Bucket returns the configured bucket name.
func (s *S3Client) Bucket() string {
	if s == nil {
		return ""
	}
	return s.bucket
}

// AvatarURL builds the proxied public URL for an avatar object key.
// The URL points to the backend's /media/{bucket}/{key} proxy endpoint.
func (s *S3Client) AvatarURL(baseURL, objectKey string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf("%s/api/v1/media/%s/%s", base, s.bucket, objectKey)
}

