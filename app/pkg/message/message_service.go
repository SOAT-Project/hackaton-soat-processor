package message

import "context"

type MessageService interface {
	SendMessage(ctx context.Context, queueURL string, messageBody string) (string, error)
}
