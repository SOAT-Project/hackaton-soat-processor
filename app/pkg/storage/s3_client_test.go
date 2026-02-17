package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func TestS3Client_Implementation(t *testing.T) {
	// Verifica se S3Client implementa a interface StorageService
	var _ StorageService = (*S3Client)(nil)
}

func TestNewS3Client(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	client := NewS3Client(cfg)

	if client == nil {
		t.Fatal("NewS3Client returned nil")
	}

	if client.client == nil {
		t.Error("S3Client.client is nil")
	}
}

func TestMockS3Service_GetObject(t *testing.T) {
	ctx := context.Background()
	expectedContent := "test content"

	mock := &MockS3Service{
		GetObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(expectedContent)), nil
		},
	}

	result, err := mock.GetObject(ctx, "test-bucket", "test-key")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	defer result.Close()

	content, err := io.ReadAll(result)
	if err != nil {
		t.Fatalf("Failed to read content: %v", err)
	}

	if string(content) != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, string(content))
	}
}

func TestMockS3Service_PutObject(t *testing.T) {
	ctx := context.Background()
	expectedKey := "test-key"

	mock := &MockS3Service{
		PutObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			return key, nil
		},
	}

	body := bytes.NewReader([]byte("test content"))
	resultKey, err := mock.PutObject(ctx, "test-bucket", expectedKey, body)

	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	if resultKey != expectedKey {
		t.Errorf("Expected key %q, got %q", expectedKey, resultKey)
	}
}

// Teste de integração básico (requer configuração AWS válida)
func TestS3Client_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Este teste só deve ser executado com credenciais AWS válidas
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Skip("AWS config not available, skipping integration test")
	}

	client := NewS3Client(cfg)
	if client == nil {
		t.Fatal("Failed to create S3Client")
	}

	// Nota: Para executar este teste completo, você precisaria:
	// 1. Ter credenciais AWS configuradas
	// 2. Ter um bucket de teste
	// 3. Implementar os testes de PutObject e GetObject reais
	t.Log("S3Client created successfully for integration testing")
}
