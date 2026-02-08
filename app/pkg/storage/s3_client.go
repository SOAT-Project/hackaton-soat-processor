package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client implementa a interface StorageService usando o AWS SDK para S3
type S3Client struct {
	client *s3.Client
}

// NewS3Client cria uma nova instância do S3Client
func NewS3Client(cfg aws.Config) *S3Client {
	return &S3Client{
		client: s3.NewFromConfig(cfg),
	}
}

// GetObject recupera um objeto do S3 a partir de sua key
func (s *S3Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	return result.Body, nil
}

// PutObject persiste um objeto no S3 e retorna sua key
func (s *S3Client) PutObject(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to put object to S3: %w", err)
	}

	return key, nil
}

// DeleteObject remove um objeto do S3
func (s *S3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}
