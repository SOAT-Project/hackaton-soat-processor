package message

import "context"

// MockMessageService é um mock da interface MessageService para testes
type MockMessageService struct {
	SendMessageFunc func(ctx context.Context, queueURL string, messageBody string) (string, error)
}

// SendMessage implementa MessageService.SendMessage usando a função mock configurada
func (m *MockMessageService) SendMessage(ctx context.Context, queueURL string, messageBody string) (string, error) {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(ctx, queueURL, messageBody)
	}
	return "mock-message-id", nil
}
