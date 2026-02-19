package adapter

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

// Mock StorageService
type mockStorageService struct {
	getObjectFunc    func(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	putObjectFunc    func(ctx context.Context, bucket, key string, body io.Reader) (string, error)
	deleteObjectFunc func(ctx context.Context, bucket, key string) error
}

func (m *mockStorageService) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, bucket, key)
	}
	return nil, nil
}

func (m *mockStorageService) PutObject(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, bucket, key, body)
	}
	return "", nil
}

func (m *mockStorageService) DeleteObject(ctx context.Context, bucket, key string) error {
	if m.deleteObjectFunc != nil {
		return m.deleteObjectFunc(ctx, bucket, key)
	}
	return nil
}

func TestNewStorageAdapter(t *testing.T) {
	mock := &mockStorageService{}
	adapter := NewStorageAdapter(mock)

	if adapter == nil {
		t.Fatal("NewStorageAdapter returned nil")
	}
}

func TestStorageAdapter_GetObject_Success(t *testing.T) {
	mock := &mockStorageService{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("test content")), nil
		},
	}

	adapter := NewStorageAdapter(mock)
	ctx := context.Background()

	reader, err := adapter.GetObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if reader == nil {
		t.Fatal("GetObject returned nil reader")
	}

	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read content: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Expected 'test content', got %s", string(content))
	}
}

func TestStorageAdapter_GetObject_Error(t *testing.T) {
	expectedError := errors.New("storage error")
	mock := &mockStorageService{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			return nil, expectedError
		},
	}

	adapter := NewStorageAdapter(mock)
	ctx := context.Background()

	_, err := adapter.GetObject(ctx, "test-bucket", "test-key")
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestStorageAdapter_PutObject_Success(t *testing.T) {
	expectedLocation := "s3://bucket/key"
	mock := &mockStorageService{
		putObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			return expectedLocation, nil
		},
	}

	adapter := NewStorageAdapter(mock)
	ctx := context.Background()
	body := strings.NewReader("upload content")

	location, err := adapter.PutObject(ctx, "test-bucket", "test-key", body)
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}
	if location != expectedLocation {
		t.Errorf("Expected location %s, got %s", expectedLocation, location)
	}
}

func TestStorageAdapter_PutObject_Error(t *testing.T) {
	expectedError := errors.New("upload error")
	mock := &mockStorageService{
		putObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			return "", expectedError
		},
	}

	adapter := NewStorageAdapter(mock)
	ctx := context.Background()
	body := strings.NewReader("upload content")

	_, err := adapter.PutObject(ctx, "test-bucket", "test-key", body)
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestStorageAdapter_DeleteObject_Success(t *testing.T) {
	mock := &mockStorageService{
		deleteObjectFunc: func(ctx context.Context, bucket, key string) error {
			return nil
		},
	}

	adapter := NewStorageAdapter(mock)
	ctx := context.Background()

	err := adapter.DeleteObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}
}

func TestStorageAdapter_DeleteObject_Error(t *testing.T) {
	expectedError := errors.New("delete error")
	mock := &mockStorageService{
		deleteObjectFunc: func(ctx context.Context, bucket, key string) error {
			return expectedError
		},
	}

	adapter := NewStorageAdapter(mock)
	ctx := context.Background()

	err := adapter.DeleteObject(ctx, "test-bucket", "test-key")
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestStorageAdapter_AllOperations(t *testing.T) {
	mock := &mockStorageService{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			if bucket == "" || key == "" {
				return nil, errors.New("invalid parameters")
			}
			return io.NopCloser(strings.NewReader("data")), nil
		},
		putObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			if bucket == "" || key == "" {
				return "", errors.New("invalid parameters")
			}
			return "location", nil
		},
		deleteObjectFunc: func(ctx context.Context, bucket, key string) error {
			if bucket == "" || key == "" {
				return errors.New("invalid parameters")
			}
			return nil
		},
	}

	adapter := NewStorageAdapter(mock)
	ctx := context.Background()

	// Test GetObject
	reader, err := adapter.GetObject(ctx, "bucket", "key")
	if err != nil {
		t.Errorf("GetObject failed: %v", err)
	}
	if reader != nil {
		reader.Close()
	}

	// Test PutObject
	_, err = adapter.PutObject(ctx, "bucket", "key", strings.NewReader("data"))
	if err != nil {
		t.Errorf("PutObject failed: %v", err)
	}

	// Test DeleteObject
	err = adapter.DeleteObject(ctx, "bucket", "key")
	if err != nil {
		t.Errorf("DeleteObject failed: %v", err)
	}
}
