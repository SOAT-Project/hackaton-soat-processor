package adapter

import (
	"context"

	"github.com/SOAT-Project/hackaton-soat-processor/internal/port"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/message"
)

type MessageAdapter struct {
	service message.MessageService
}

func NewMessageAdapter(service message.MessageService) port.MessagePort {
	return &MessageAdapter{
		service: service,
	}
}

func (a *MessageAdapter) SendMessage(ctx context.Context, queueURL string, messageBody string) (string, error) {
	return a.service.SendMessage(ctx, queueURL, messageBody)
}
