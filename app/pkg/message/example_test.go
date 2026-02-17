package message_test

import (
	"context"
	"fmt"
	"log"

	"github.com/SOAT-Project/hackaton-soat-processor/pkg/message"
	"github.com/aws/aws-sdk-go-v2/config"
)

// ExampleSQSClient_SendMessage demonstra como enviar uma mensagem para o SQS
func ExampleSQSClient_SendMessage() {
	ctx := context.Background()

	// Carrega a configuração AWS
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	// Cria o cliente SQS
	sqsService := message.NewSQSClient(cfg)

	// Envia uma mensagem
	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789/my-queue"
	messageBody := `{"order_id": "12345", "status": "completed"}`

	messageID, err := sqsService.SendMessage(ctx, queueURL, messageBody)
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	fmt.Printf("Message sent successfully with ID: %s\n", messageID)
}

// ExampleMockMessageService demonstra como usar o mock para testes
func ExampleMockMessageService() {
	ctx := context.Background()

	// Cria um mock do serviço de mensagens
	mockSQS := &message.MockMessageService{
		SendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			// Simula o envio bem-sucedido
			fmt.Printf("Mock: Sending message to %s\n", queueURL)
			fmt.Printf("Mock: Message body: %s\n", messageBody)
			return "mock-message-id-123", nil
		},
	}

	// Usa o mock como se fosse o serviço real
	var sqsService message.MessageService = mockSQS

	// Envia uma mensagem usando o mock
	messageID, err := sqsService.SendMessage(
		ctx,
		"https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
		`{"test": "data"}`,
	)
	if err != nil {
		log.Fatalf("failed to send message: %v", err)
	}

	fmt.Printf("Message ID: %s\n", messageID)
}
