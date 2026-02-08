package storage

import (
	"context"
	"io"
)

// MockS3Service é um mock da interface StorageService para testes
type MockS3Service struct {
	GetObjectFunc    func(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	PutObjectFunc    func(ctx context.Context, bucket, key string, body io.Reader) (string, error)
	DeleteObjectFunc func(ctx context.Context, bucket, key string) error
}

// GetObject implementa StorageService.GetObject usando a função mock configurada
func (m *MockS3Service) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if m.GetObjectFunc != nil {
		return m.GetObjectFunc(ctx, bucket, key)
	}
	return nil, nil
}

// PutObject implementa StorageService.PutObject usando a função mock configurada
func (m *MockS3Service) PutObject(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
	if m.PutObjectFunc != nil {
		return m.PutObjectFunc(ctx, bucket, key, body)
	}
	return key, nil
}

// DeleteObject implementa StorageService.DeleteObject usando a função mock configurada
func (m *MockS3Service) DeleteObject(ctx context.Context, bucket, key string) error {
	if m.DeleteObjectFunc != nil {
		return m.DeleteObjectFunc(ctx, bucket, key)
	}
	return nil
}
