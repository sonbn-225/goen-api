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

type S3Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseSSL        bool
	PublicBaseURL string
}

type S3Client struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

func NewS3Client(cfg S3Config) (*S3Client, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, nil // Silently skip if no endpoint (standard dev behavior)
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: "auto", // Most compatible with SeaweedFS/S3 variants
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3 client at %s: %w", cfg.Endpoint, err)
	}

	s := &S3Client{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: cfg.PublicBaseURL,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket %s: %w", cfg.Bucket, err)
	}
	
	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.Bucket, err)
		}
	}

	return s, nil
}

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

func (s *S3Client) ListObjects(ctx context.Context, prefix string) ([]minio.ObjectInfo, error) {
	if s == nil {
		return nil, fmt.Errorf("storage not configured")
	}

	objects := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var result []minio.ObjectInfo
	for obj := range objects {
		if obj.Err != nil {
			return nil, obj.Err
		}
		result = append(result, obj)
	}
	return result, nil
}

type S3ObjectInfo struct {
	Size        int64
	ContentType string
}

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

	return obj, S3ObjectInfo{
		Size:        stat.Size,
		ContentType: stat.ContentType,
	}, nil
}

func (s *S3Client) AvatarURL(baseURL, objectKey string) string {
	base := strings.TrimRight(baseURL, "/")
	key := strings.TrimLeft(objectKey, "/")
	if base == "" {
		return fmt.Sprintf("/api/v1/media/%s/%s", s.bucket, key)
	}
	return fmt.Sprintf("%s/api/v1/media/%s/%s", base, s.bucket, key)
}

func (s *S3Client) Bucket() string {
	if s == nil {
		return ""
	}
	return s.bucket
}
