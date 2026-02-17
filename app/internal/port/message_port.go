package port

import "context"

type MessagePort interface {
	SendMessage(ctx context.Context, queueURL string, messageBody string) (string, error)
}
