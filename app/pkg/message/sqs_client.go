package message

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSClient implementa a interface MessageService usando o AWS SQS
type SQSClient struct {
	client *sqs.Client
}

// NewSQSClient cria uma nova instância do SQSClient
func NewSQSClient(cfg aws.Config) *SQSClient {
	return &SQSClient{
		client: sqs.NewFromConfig(cfg),
	}
}

// SendMessage envia uma mensagem para uma fila SQS
func (s *SQSClient) SendMessage(ctx context.Context, queueURL string, messageBody string) (string, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(messageBody),
	}

	result, err := s.client.SendMessage(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to send message to SQS: %w", err)
	}

	if result.MessageId == nil {
		return "", fmt.Errorf("message sent but no message ID returned")
	}

	return *result.MessageId, nil
}
