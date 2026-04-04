package objectstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config holds SeaweedFS or S3-compatible object storage settings.
type Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseSSL        bool
	PublicBaseURL string
}

// ObjectInfo contains object metadata.
type ObjectInfo struct {
	Size        int64
	ContentType string
}

// SeaweedClient is an S3-compatible client used for avatar storage.
type SeaweedClient struct {
	client        *minio.Client
	bucket        string
	endpoint      string
	useSSL        bool
	publicBaseURL string
}

func NewSeaweedClient(cfg Config) *SeaweedClient {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		return nil
	}

	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" {
		bucket = "goen"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil
	}

	s := &SeaweedClient{
		client:        client,
		bucket:        bucket,
		endpoint:      endpoint,
		useSSL:        cfg.UseSSL,
		publicBaseURL: strings.TrimSpace(cfg.PublicBaseURL),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if exists, err := client.BucketExists(ctx, bucket); err == nil && !exists {
		_ = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}

	return s
}

func (s *SeaweedClient) UploadAvatar(ctx context.Context, userID, fileName, contentType string, data []byte) (string, error) {
	if s == nil {
		return "", fmt.Errorf("storage not configured")
	}
	if strings.TrimSpace(userID) == "" {
		return "", fmt.Errorf("user id is required")
	}

	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	if ext == "" {
		ext = extensionFromContentType(contentType)
	}
	if ext == "" {
		ext = ".bin"
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectKey := fmt.Sprintf("avatars/%s/%s%s", userID, uuid.NewString(), ext)
	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectKey,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", err
	}

	return s.AvatarURL(objectKey), nil
}

func (s *SeaweedClient) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectInfo, error) {
	if s == nil {
		return nil, ObjectInfo{}, fmt.Errorf("storage not configured")
	}
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	stat, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, ObjectInfo{}, err
	}
	return obj, ObjectInfo{Size: stat.Size, ContentType: stat.ContentType}, nil
}

func (s *SeaweedClient) Bucket() string {
	if s == nil {
		return ""
	}
	return s.bucket
}

func (s *SeaweedClient) AvatarURL(objectKey string) string {
	if s == nil {
		return ""
	}

	key := strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	path := fmt.Sprintf("/api/v1/media/%s/%s", s.bucket, key)
	if s.publicBaseURL != "" {
		return fmt.Sprintf("%s%s", strings.TrimRight(s.publicBaseURL, "/"), path)
	}
	return path
}

func extensionFromContentType(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}
