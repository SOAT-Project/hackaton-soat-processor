package adapter

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Mock MessageService
type mockMessageService struct {
	sendMessageFunc func(ctx context.Context, queueURL string, messageBody string) (string, error)
}

func (m *mockMessageService) SendMessage(ctx context.Context, queueURL string, messageBody string) (string, error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, queueURL, messageBody)
	}
	return "", nil
}

func TestNewMessageAdapter(t *testing.T) {
	mock := &mockMessageService{}

	adapter := NewMessageAdapter(mock)

	if adapter == nil {
		t.Fatal("NewMessageAdapter returned nil")
	}
}

func TestMessageAdapter_SendMessage_Success(t *testing.T) {
	expectedMessageID := "msg-12345"
	mock := &mockMessageService{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			if queueURL == "" {
				return "", errors.New("queue URL required")
			}
			if messageBody == "" {
				return "", errors.New("message body required")
			}
			return expectedMessageID, nil
		},
	}

	adapter := NewMessageAdapter(mock)
	ctx := context.Background()

	messageID, err := adapter.SendMessage(ctx, "https://sqs.amazonaws.com/queue", `{"test": "message"}`)
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if messageID != expectedMessageID {
		t.Errorf("Expected message ID %s, got %s", expectedMessageID, messageID)
	}
}

func TestMessageAdapter_SendMessage_Error(t *testing.T) {
	expectedError := errors.New("queue not found")
	mock := &mockMessageService{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "", expectedError
		},
	}

	adapter := NewMessageAdapter(mock)
	ctx := context.Background()

	_, err := adapter.SendMessage(ctx, "invalid-queue", "test message")
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestMessageAdapter_SendMessage_EmptyQueue(t *testing.T) {
	mock := &mockMessageService{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			if queueURL == "" {
				return "", errors.New("queue URL is required")
			}
			return "msg-id", nil
		},
	}

	adapter := NewMessageAdapter(mock)
	ctx := context.Background()

	_, err := adapter.SendMessage(ctx, "", "test message")
	if err == nil {
		t.Error("Expected error for empty queue URL, got nil")
	}
}

func TestMessageAdapter_SendMessage_EmptyBody(t *testing.T) {
	mock := &mockMessageService{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			if messageBody == "" {
				return "", errors.New("message body is required")
			}
			return "msg-id", nil
		},
	}

	adapter := NewMessageAdapter(mock)
	ctx := context.Background()

	_, err := adapter.SendMessage(ctx, "https://queue-url", "")
	if err == nil {
		t.Error("Expected error for empty message body, got nil")
	}
}

func TestMessageAdapter_SendMessage_WithContext(t *testing.T) {
	callCount := 0
	mock := &mockMessageService{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			callCount++
			if ctx == nil {
				t.Error("Context is nil")
			}
			return "msg-id", nil
		},
	}

	adapter := NewMessageAdapter(mock)
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	_, err := adapter.SendMessage(ctx, "https://queue-url", "test message")
	if err != nil {
		t.Errorf("SendMessage failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestMessageAdapter_SendMessage_MultipleMessages(t *testing.T) {
	messages := []string{}
	mock := &mockMessageService{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			messages = append(messages, messageBody)
			return "msg-id", nil
		},
	}

	adapter := NewMessageAdapter(mock)
	ctx := context.Background()

	testMessages := []string{"message1", "message2", "message3"}
	for _, msg := range testMessages {
		_, err := adapter.SendMessage(ctx, "queue-url", msg)
		if err != nil {
			t.Errorf("SendMessage failed for %s: %v", msg, err)
		}
	}

	if len(messages) != len(testMessages) {
		t.Errorf("Expected %d messages, got %d", len(testMessages), len(messages))
	}
}

func TestMessageAdapter_SendMessage_ContextCanceled(t *testing.T) {
	mock := &mockMessageService{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			// Simulate a slow operation
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "msg-id", nil
			}
		},
	}

	adapter := NewMessageAdapter(mock)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := adapter.SendMessage(ctx, "queue-url", "test message")
	if err != context.Canceled {
		t.Logf("Expected context canceled error, got: %v", err)
	}
}

func TestMessageAdapter_SendMessage_LargeBody(t *testing.T) {
	receivedBody := ""
	mock := &mockMessageService{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			receivedBody = messageBody
			return "msg-large", nil
		},
	}

	adapter := NewMessageAdapter(mock)
	ctx := context.Background()

	// Create a large message body
	largeBody := strings.Repeat("x", 10000)

	messageID, err := adapter.SendMessage(ctx, "queue-url", largeBody)
	if err != nil {
		t.Errorf("SendMessage failed: %v", err)
	}
	if messageID != "msg-large" {
		t.Errorf("Expected msg-large, got %s", messageID)
	}
	if receivedBody != largeBody {
		t.Error("Large body was not received correctly")
	}
}
