package message

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func TestSQSClient_Implementation(t *testing.T) {
	// Verifica se SQSClient implementa a interface MessageService
	var _ MessageService = (*SQSClient)(nil)
}

func TestNewSQSClient(t *testing.T) {
	cfg := aws.Config{
		Region: "us-east-1",
	}

	client := NewSQSClient(cfg)

	if client == nil {
		t.Fatal("NewSQSClient returned nil")
	}

	if client.client == nil {
		t.Error("SQSClient.client is nil")
	}
}

func TestMockMessageService_SendMessage(t *testing.T) {
	ctx := context.Background()
	expectedMessageID := "test-message-id-123"
	expectedQueueURL := "https://sqs.us-east-1.amazonaws.com/123456789/test-queue"
	expectedBody := "test message body"

	mock := &MockMessageService{
		SendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			if queueURL != expectedQueueURL {
				t.Errorf("Expected queue URL %q, got %q", expectedQueueURL, queueURL)
			}
			if messageBody != expectedBody {
				t.Errorf("Expected message body %q, got %q", expectedBody, messageBody)
			}
			return expectedMessageID, nil
		},
	}

	messageID, err := mock.SendMessage(ctx, expectedQueueURL, expectedBody)
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if messageID != expectedMessageID {
		t.Errorf("Expected message ID %q, got %q", expectedMessageID, messageID)
	}
}

func TestMockMessageService_DefaultBehavior(t *testing.T) {
	ctx := context.Background()

	// Mock sem função configurada deve retornar valor padrão
	mock := &MockMessageService{}

	messageID, err := mock.SendMessage(ctx, "test-queue", "test message")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if messageID != "mock-message-id" {
		t.Errorf("Expected default message ID 'mock-message-id', got %q", messageID)
	}
}

// Teste de integração básico (requer configuração AWS válida)
func TestSQSClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Este teste só deve ser executado com credenciais AWS válidas
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Skip("AWS config not available, skipping integration test")
	}

	client := NewSQSClient(cfg)
	if client == nil {
		t.Fatal("Failed to create SQSClient")
	}

	// Nota: Para executar este teste completo, você precisaria:
	// 1. Ter credenciais AWS configuradas
	// 2. Ter uma fila SQS de teste
	// 3. Implementar o teste de SendMessage real
	t.Log("SQSClient created successfully for integration testing")
}
