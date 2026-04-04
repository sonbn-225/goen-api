package media

import (
	"context"
	"io"
)

type ObjectInfo struct {
	ContentType string
}

type Storage interface {
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectInfo, error)
}

type Service interface {
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, ObjectInfo, error)
}

type ModuleDeps struct {
	Storage Storage
	Service Service
}

type Module struct {
	Service Service
	Handler *Handler
}
