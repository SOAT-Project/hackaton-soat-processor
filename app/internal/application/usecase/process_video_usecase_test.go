package usecase

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/SOAT-Project/hackaton-soat-processor/internal/application/domain"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/observability"
)

// Mock implementations for testing

type mockStoragePort struct {
	getObjectFunc    func(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	putObjectFunc    func(ctx context.Context, bucket, key string, body io.Reader) (string, error)
	deleteObjectFunc func(ctx context.Context, bucket, key string) error
}

func (m *mockStoragePort) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, bucket, key)
	}
	return io.NopCloser(strings.NewReader("mock video data")), nil
}

func (m *mockStoragePort) PutObject(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, bucket, key, body)
	}
	return key, nil
}

func (m *mockStoragePort) DeleteObject(ctx context.Context, bucket, key string) error {
	if m.deleteObjectFunc != nil {
		return m.deleteObjectFunc(ctx, bucket, key)
	}
	return nil
}

type mockMessagePort struct {
	sendMessageFunc func(ctx context.Context, queueURL string, messageBody string) (string, error)
}

func (m *mockMessagePort) SendMessage(ctx context.Context, queueURL string, messageBody string) (string, error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, queueURL, messageBody)
	}
	return "mock-message-id", nil
}

type mockVideoProcessor struct {
	processVideoFunc func(ctx context.Context, videoPath string) (string, int, error)
}

func (m *mockVideoProcessor) ProcessVideo(ctx context.Context, videoPath string) (string, int, error) {
	if m.processVideoFunc != nil {
		return m.processVideoFunc(ctx, videoPath)
	}
	return "/tmp/mock.zip", 10, nil
}

func TestNewProcessVideoUseCase(t *testing.T) {
	storage := &mockStoragePort{}
	message := &mockMessagePort{}
	processor := &mockVideoProcessor{}

	useCase := NewProcessVideoUseCase(storage, message, processor, "test-bucket", "test-queue")

	if useCase == nil {
		t.Fatal("Expected use case to be created")
	}

	if useCase.outputBucket != "test-bucket" {
		t.Errorf("Expected output bucket 'test-bucket', got '%s'", useCase.outputBucket)
	}

	if useCase.outputQueueURL != "test-queue" {
		t.Errorf("Expected output queue 'test-queue', got '%s'", useCase.outputQueueURL)
	}
}

func TestValidateRequest(t *testing.T) {
	useCase := NewProcessVideoUseCase(nil, nil, nil, "", "")

	tests := []struct {
		name    string
		request domain.VideoProcess
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: domain.VideoProcess{
				ProcessID:   "123",
				VideoBucket: "test-bucket",
				VideoKey:    "video.mp4",
			},
			wantErr: false,
		},
		{
			name: "missing process_id",
			request: domain.VideoProcess{
				VideoBucket: "test-bucket",
				VideoKey:    "video.mp4",
			},
			wantErr: true,
			errMsg:  "process_id is required",
		},
		{
			name: "missing video_bucket",
			request: domain.VideoProcess{
				ProcessID: "123",
				VideoKey:  "video.mp4",
			},
			wantErr: true,
			errMsg:  "video_bucket is required",
		},
		{
			name: "missing video_key",
			request: domain.VideoProcess{
				ProcessID:   "123",
				VideoBucket: "test-bucket",
			},
			wantErr: true,
			errMsg:  "video_key is required",
		},
		{
			name: "invalid video format",
			request: domain.VideoProcess{
				ProcessID:   "123",
				VideoBucket: "test-bucket",
				VideoKey:    "document.pdf",
			},
			wantErr: true,
			errMsg:  "invalid video file format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.validateRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateRequest() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestIsValidVideoFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"video.mp4", true},
		{"video.avi", true},
		{"video.mov", true},
		{"video.mkv", true},
		{"video.wmv", true},
		{"video.flv", true},
		{"video.webm", true},
		{"document.pdf", false},
		{"image.jpg", false},
		{"file.txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := isValidVideoFile(tt.filename)
			if got != tt.want {
				t.Errorf("isValidVideoFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestExecute_ValidationError(t *testing.T) {
	var sentMessage string
	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			sentMessage = messageBody
			return "msg-id", nil
		},
	}

	useCase := NewProcessVideoUseCase(nil, messagePort, nil, "test-bucket", "test-queue")

	request := domain.VideoProcess{
		ProcessID: "", // Invalid: empty process_id
	}

	err := useCase.Execute(context.Background(), request)
	if err == nil {
		t.Fatal("Expected error for invalid request")
	}

	if sentMessage == "" {
		t.Error("Expected error message to be sent to queue")
	}

	if !strings.Contains(sentMessage, "error_message") {
		t.Errorf("Expected error message format, got: %s", sentMessage)
	}
}

func TestExecute_StorageError(t *testing.T) {
	storagePort := &mockStoragePort{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			return nil, errors.New("storage error")
		},
	}

	var sentMessage string
	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			sentMessage = messageBody
			return "msg-id", nil
		},
	}

	useCase := NewProcessVideoUseCase(storagePort, messagePort, nil, "test-bucket", "test-queue")

	request := domain.VideoProcess{
		ProcessID:   "123",
		VideoBucket: "test-bucket",
		VideoKey:    "video.mp4",
	}

	err := useCase.Execute(context.Background(), request)
	if err == nil {
		t.Fatal("Expected error when storage fails")
	}

	if !strings.Contains(sentMessage, "error_message") {
		t.Errorf("Expected error message to be sent, got: %s", sentMessage)
	}
}

func TestExecute_FullSuccess(t *testing.T) {
	// Initialize logger
	if err := observability.InitLogger("test"); err != nil {
		t.Fatalf("Failed to setup logger: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "test-video-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.WriteString("fake video content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	zipFile, err := os.CreateTemp("", "test-zip-*.zip")
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	zipFile.WriteString("fake zip content")
	zipFile.Close()
	defer os.Remove(zipFile.Name())

	storagePort := &mockStoragePort{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			file, err := os.Open(tmpFile.Name())
			return file, err
		},
		putObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			return "s3://bucket/key", nil
		},
		deleteObjectFunc: func(ctx context.Context, bucket, key string) error {
			return nil
		},
	}

	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "msg-success", nil
		},
	}

	videoProcessor := &mockVideoProcessor{
		processVideoFunc: func(ctx context.Context, videoPath string) (string, int, error) {
			return zipFile.Name(), 30, nil
		},
	}

	useCase := NewProcessVideoUseCase(storagePort, messagePort, videoProcessor, "output-bucket", "output-queue")

	request := domain.VideoProcess{
		ProcessID:   "process-123",
		VideoBucket: "input-bucket",
		VideoKey:    "video.mp4",
	}

	err = useCase.Execute(context.Background(), request)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestExecute_ProcessingError(t *testing.T) {
	if err := observability.InitLogger("test"); err != nil {
		t.Fatalf("Failed to setup logger: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "test-video-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.WriteString("fake video content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storagePort := &mockStoragePort{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			file, err := os.Open(tmpFile.Name())
			return file, err
		},
	}

	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "msg-id", nil
		},
	}

	videoProcessor := &mockVideoProcessor{
		processVideoFunc: func(ctx context.Context, videoPath string) (string, int, error) {
			return "", 0, errors.New("processing failed")
		},
	}

	useCase := NewProcessVideoUseCase(storagePort, messagePort, videoProcessor, "output-bucket", "output-queue")

	request := domain.VideoProcess{
		ProcessID:   "process-456",
		VideoBucket: "input-bucket",
		VideoKey:    "video.mp4",
	}

	err = useCase.Execute(context.Background(), request)
	if err == nil {
		t.Fatal("Expected error from processing")
	}
}

func TestExecute_UploadError(t *testing.T) {
	if err := observability.InitLogger("test"); err != nil {
		t.Fatalf("Failed to setup logger: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "test-video-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.WriteString("fake video content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	zipFile, err := os.CreateTemp("", "test-zip-*.zip")
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	zipFile.WriteString("fake zip content")
	zipFile.Close()
	defer os.Remove(zipFile.Name())

	storagePort := &mockStoragePort{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			file, err := os.Open(tmpFile.Name())
			return file, err
		},
		putObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			return "", errors.New("upload failed")
		},
	}

	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "msg-id", nil
		},
	}

	videoProcessor := &mockVideoProcessor{
		processVideoFunc: func(ctx context.Context, videoPath string) (string, int, error) {
			return zipFile.Name(), 25, nil
		},
	}

	useCase := NewProcessVideoUseCase(storagePort, messagePort, videoProcessor, "output-bucket", "output-queue")

	request := domain.VideoProcess{
		ProcessID:   "process-789",
		VideoBucket: "input-bucket",
		VideoKey:    "video.mp4",
	}

	err = useCase.Execute(context.Background(), request)
	if err == nil {
		t.Fatal("Expected error from upload")
	}
}

func TestExecute_DeleteOriginalVideoError(t *testing.T) {
	if err := observability.InitLogger("test"); err != nil {
		t.Fatalf("Failed to setup logger: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "test-video-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.WriteString("fake video content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	zipFile, err := os.CreateTemp("", "test-zip-*.zip")
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	zipFile.WriteString("fake zip content")
	zipFile.Close()
	defer os.Remove(zipFile.Name())

	storagePort := &mockStoragePort{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			file, err := os.Open(tmpFile.Name())
			return file, err
		},
		putObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			return "s3://bucket/key", nil
		},
		deleteObjectFunc: func(ctx context.Context, bucket, key string) error {
			return errors.New("delete failed")
		},
	}

	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "msg-success", nil
		},
	}

	videoProcessor := &mockVideoProcessor{
		processVideoFunc: func(ctx context.Context, videoPath string) (string, int, error) {
			return zipFile.Name(), 20, nil
		},
	}

	useCase := NewProcessVideoUseCase(storagePort, messagePort, videoProcessor, "output-bucket", "output-queue")

	request := domain.VideoProcess{
		ProcessID:   "process-delete",
		VideoBucket: "input-bucket",
		VideoKey:    "video.mp4",
	}

	// Should succeed even if delete fails
	err = useCase.Execute(context.Background(), request)
	if err != nil {
		t.Fatalf("Execute should succeed even with delete error: %v", err)
	}
}

func TestExecute_SendSuccessMessageError(t *testing.T) {
	if err := observability.InitLogger("test"); err != nil {
		t.Fatalf("Failed to setup logger: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "test-video-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.WriteString("fake video content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	zipFile, err := os.CreateTemp("", "test-zip-*.zip")
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	zipFile.WriteString("fake zip content")
	zipFile.Close()
	defer os.Remove(zipFile.Name())

	storagePort := &mockStoragePort{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			file, err := os.Open(tmpFile.Name())
			return file, err
		},
		putObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			return "s3://bucket/key", nil
		},
		deleteObjectFunc: func(ctx context.Context, bucket, key string) error {
			return nil
		},
	}

	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "", errors.New("send message failed")
		},
	}

	videoProcessor := &mockVideoProcessor{
		processVideoFunc: func(ctx context.Context, videoPath string) (string, int, error) {
			return zipFile.Name(), 15, nil
		},
	}

	useCase := NewProcessVideoUseCase(storagePort, messagePort, videoProcessor, "output-bucket", "output-queue")

	request := domain.VideoProcess{
		ProcessID:   "process-msg-error",
		VideoBucket: "input-bucket",
		VideoKey:    "video.mp4",
	}

	err = useCase.Execute(context.Background(), request)
	if err == nil {
		t.Fatal("Expected error from send message")
	}
}

func TestExecute_SendErrorMessage_ValidationError(t *testing.T) {
	if err := observability.InitLogger("test"); err != nil {
		t.Fatalf("Failed to setup logger: %v", err)
	}

	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "msg-id", nil
		},
	}

	useCase := NewProcessVideoUseCase(nil, messagePort, nil, "output-bucket", "output-queue")

	// Test with empty request to trigger validation error
	request := domain.VideoProcess{
		ProcessID:   "",
		VideoBucket: "",
		VideoKey:    "",
	}

	err := useCase.Execute(context.Background(), request)
	if err == nil {
		t.Fatal("Expected validation error")
	}
}

func TestDownloadVideo_ReadError(t *testing.T) {
	if err := observability.InitLogger("test"); err != nil {
		t.Fatalf("Failed to setup logger: %v", err)
	}

	// Create a reader that will fail on read
	failingReader := &failingReadCloser{}

	storagePort := &mockStoragePort{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			return failingReader, nil
		},
	}

	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "msg-id", nil
		},
	}

	useCase := NewProcessVideoUseCase(storagePort, messagePort, nil, "output-bucket", "output-queue")

	request := domain.VideoProcess{
		ProcessID:   "test-fail-read",
		VideoBucket: "bucket",
		VideoKey:    "video.mp4",
	}

	err := useCase.Execute(context.Background(), request)
	if err == nil {
		t.Fatal("Expected error from read failure")
	}
}

// Helper type for testing read failures
type failingReadCloser struct{}

func (f *failingReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (f *failingReadCloser) Close() error {
	return nil
}

func TestUploadZip_OpenFileError(t *testing.T) {
	if err := observability.InitLogger("test"); err != nil {
		t.Fatalf("Failed to setup logger: %v", err)
	}

	tmpFile, err := os.CreateTemp("", "test-video-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.WriteString("fake video content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	zipFile, err := os.CreateTemp("", "test-zip-*.zip")
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	zipFile.WriteString("fake zip content")
	zipFile.Close()
	zipPath := zipFile.Name()
	// Remove it to cause open error
	os.Remove(zipPath)

	storagePort := &mockStoragePort{
		getObjectFunc: func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
			file, err := os.Open(tmpFile.Name())
			return file, err
		},
		putObjectFunc: func(ctx context.Context, bucket, key string, body io.Reader) (string, error) {
			return "s3://bucket/key", nil
		},
		deleteObjectFunc: func(ctx context.Context, bucket, key string) error {
			return nil
		},
	}

	messagePort := &mockMessagePort{
		sendMessageFunc: func(ctx context.Context, queueURL string, messageBody string) (string, error) {
			return "msg-success", nil
		},
	}

	videoProcessor := &mockVideoProcessor{
		processVideoFunc: func(ctx context.Context, videoPath string) (string, int, error) {
			// Return the removed zip path to trigger open error
			return zipPath, 10, nil
		},
	}

	useCase := NewProcessVideoUseCase(storagePort, messagePort, videoProcessor, "output-bucket", "output-queue")

	request := domain.VideoProcess{
		ProcessID:   "test-open-error",
		VideoBucket: "input-bucket",
		VideoKey:    "video.mp4",
	}

	err = useCase.Execute(context.Background(), request)
	if err == nil {
		t.Fatal("Expected error from file open")
	}
}
