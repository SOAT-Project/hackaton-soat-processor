package adapter

import (
	"context"
	"io"

	"github.com/SOAT-Project/hackaton-soat-processor/internal/port"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/storage"
)

type StorageAdapter struct {
	service storage.StorageService
}

func NewStorageAdapter(service storage.StorageService) port.StoragePort {
	return &StorageAdapter{
		service: service,
	}
}

func (a *StorageAdapter) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return a.service.GetObject(ctx, bucket, key)
}

func (a *StorageAdapter) PutObject(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
	return a.service.PutObject(ctx, bucket, key, body)
}

func (a *StorageAdapter) DeleteObject(ctx context.Context, bucket, key string) error {
	return a.service.DeleteObject(ctx, bucket, key)
}
