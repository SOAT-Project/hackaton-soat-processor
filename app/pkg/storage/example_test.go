package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/SOAT-Project/hackaton-soat-processor/pkg/storage"
	"github.com/aws/aws-sdk-go-v2/config"
)

// ExampleS3Client_GetObject demonstra como recuperar um objeto do S3
func ExampleS3Client_GetObject() {
	ctx := context.Background()

	// Carrega a configuração AWS
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	// Cria o cliente S3
	s3Service := storage.NewS3Client(cfg)

	// Recupera um objeto
	bucket := "my-bucket"
	key := "path/to/my-object.txt"

	body, err := s3Service.GetObject(ctx, bucket, key)
	if err != nil {
		log.Fatalf("failed to get object: %v", err)
	}
	defer body.Close()

	// Lê o conteúdo
	content, err := io.ReadAll(body)
	if err != nil {
		log.Fatalf("failed to read object content: %v", err)
	}

	fmt.Printf("Object content: %s\n", string(content))
}

// ExampleS3Client_PutObject demonstra como persistir um objeto no S3
func ExampleS3Client_PutObject() {
	ctx := context.Background()

	// Carrega a configuração AWS
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	// Cria o cliente S3
	s3Service := storage.NewS3Client(cfg)

	// Prepara o conteúdo a ser enviado
	content := []byte("Hello, S3!")
	body := bytes.NewReader(content)

	// Persiste o objeto
	bucket := "my-bucket"
	key := "path/to/my-new-object.txt"

	resultKey, err := s3Service.PutObject(ctx, bucket, key, body)
	if err != nil {
		log.Fatalf("failed to put object: %v", err)
	}

	fmt.Printf("Object uploaded successfully with key: %s\n", resultKey)
}

// ExampleMockS3Service demonstra como usar o mock para testes
func ExampleMockS3Service() {
	ctx := context.Background()

	// Cria um mock do serviço S3
	mockS3 := &storage.MockS3Service{
		GetObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			// Simula a leitura de um arquivo
			content := "mocked content"
			return io.NopCloser(bytes.NewReader([]byte(content))), nil
		},
		PutObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			// Simula o upload bem-sucedido
			return key, nil
		},
	}

	// Usa o mock como se fosse o serviço real
	var s3Service storage.StorageService = mockS3

	// Testa o GetObject
	body, err := s3Service.GetObject(ctx, "test-bucket", "test-key")
	if err != nil {
		log.Fatalf("failed to get object: %v", err)
	}
	defer body.Close()

	content, _ := io.ReadAll(body)
	fmt.Printf("GetObject result: %s\n", string(content))

	// Testa o PutObject
	uploadBody := bytes.NewReader([]byte("test data"))
	key, err := s3Service.PutObject(ctx, "test-bucket", "new-key", uploadBody)
	if err != nil {
		log.Fatalf("failed to put object: %v", err)
	}

	fmt.Printf("PutObject result key: %s\n", key)
}
