package storage

import (
	"context"
	"io"
)

type StorageService interface {
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	PutObject(ctx context.Context, bucket, key string, body io.Reader) (string, error)

	DeleteObject(ctx context.Context, bucket, key string) error
}
