package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/SOAT-Project/hackaton-soat-processor/internal/application/domain"
	"github.com/SOAT-Project/hackaton-soat-processor/internal/port"
)

type ProcessVideoUseCase struct {
	storage        port.StoragePort
	message        port.MessagePort
	videoProcessor port.VideoProcessorPort
	outputBucket   string
	outputQueueURL string
}

func NewProcessVideoUseCase(
	storage port.StoragePort,
	message port.MessagePort,
	videoProcessor port.VideoProcessorPort,
	outputBucket string,
	outputQueueURL string,
) *ProcessVideoUseCase {
	return &ProcessVideoUseCase{
		storage:        storage,
		message:        message,
		videoProcessor: videoProcessor,
		outputBucket:   outputBucket,
		outputQueueURL: outputQueueURL,
	}
}

func (uc *ProcessVideoUseCase) Execute(ctx context.Context, request domain.VideoProcess) error {
	log.Printf("[UseCase] Starting video processing for process_id: %s", request.ProcessID)

	result := &domain.ProcessResult{
		ProcessID: request.ProcessID,
		Success:   false,
	}

	if err := uc.validateRequest(request); err != nil {
		result.Error = err
		return uc.sendErrorMessage(ctx, result)
	}

	videoPath, err := uc.downloadVideo(ctx, request)
	if err != nil {
		result.Error = fmt.Errorf("failed to download video: %w", err)
		return uc.sendErrorMessage(ctx, result)
	}
	defer os.Remove(videoPath)

	zipPath, frameCount, err := uc.videoProcessor.ProcessVideo(ctx, videoPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to process video: %w", err)
		return uc.sendErrorMessage(ctx, result)
	}
	defer os.Remove(zipPath)

	log.Printf("[UseCase] Video processed successfully. Frames extracted: %d", frameCount)

	outputKey := fmt.Sprintf("processed/frames_%s.zip", request.ProcessID)
	if err := uc.uploadZip(ctx, zipPath, outputKey); err != nil {
		result.Error = fmt.Errorf("failed to upload zip: %w", err)
		return uc.sendErrorMessage(ctx, result)
	}

	if err := uc.deleteOriginalVideo(ctx, request); err != nil {
		log.Printf("[UseCase] Warning: failed to delete original video: %v", err)
	}

	result.Success = true
	result.FileBucket = uc.outputBucket
	result.FileKey = outputKey

	return uc.sendSuccessMessage(ctx, result)
}

func (uc *ProcessVideoUseCase) validateRequest(request domain.VideoProcess) error {
	if request.ProcessID == "" {
		return fmt.Errorf("process_id is required")
	}
	if request.VideoBucket == "" {
		return fmt.Errorf("video_bucket is required")
	}
	if request.VideoKey == "" {
		return fmt.Errorf("video_key is required")
	}

	if !isValidVideoFile(request.VideoKey) {
		return fmt.Errorf("invalid video file format. Supported: mp4, avi, mov, mkv, wmv, flv, webm")
	}

	return nil
}

func (uc *ProcessVideoUseCase) downloadVideo(ctx context.Context, request domain.VideoProcess) (string, error) {
	log.Printf("[UseCase] Downloading video from s3://%s/%s", request.VideoBucket, request.VideoKey)

	body, err := uc.storage.GetObject(ctx, request.VideoBucket, request.VideoKey)
	if err != nil {
		return "", fmt.Errorf("failed to get object from storage: %w", err)
	}
	defer body.Close()

	tempDir := "/tmp/video-processor"
	if err := os.MkdirAll(tempDir, 0777); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	ext := filepath.Ext(request.VideoKey)
	tempFile := filepath.Join(tempDir, fmt.Sprintf("video_%s%s", request.ProcessID, ext))

	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, body)
	if err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to save video: %w", err)
	}

	log.Printf("[UseCase] Video downloaded to: %s", tempFile)
	return tempFile, nil
}

func (uc *ProcessVideoUseCase) uploadZip(ctx context.Context, zipPath, outputKey string) error {
	log.Printf("[UseCase] Uploading ZIP to s3://%s/%s", uc.outputBucket, outputKey)

	file, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer file.Close()

	_, err = uc.storage.PutObject(ctx, uc.outputBucket, outputKey, file)
	if err != nil {
		return fmt.Errorf("failed to put object to storage: %w", err)
	}

	log.Printf("[UseCase] ZIP uploaded successfully")
	return nil
}

func (uc *ProcessVideoUseCase) deleteOriginalVideo(ctx context.Context, request domain.VideoProcess) error {
	log.Printf("[UseCase] Deleting original video from s3://%s/%s", request.VideoBucket, request.VideoKey)

	err := uc.storage.DeleteObject(ctx, request.VideoBucket, request.VideoKey)
	if err != nil {
		return fmt.Errorf("failed to delete original video: %w", err)
	}

	log.Printf("[UseCase] Original video deleted successfully")
	return nil
}

func (uc *ProcessVideoUseCase) sendSuccessMessage(ctx context.Context, result *domain.ProcessResult) error {
	log.Printf("[UseCase] Sending success message for process_id: %s", result.ProcessID)

	msgData := result.ToSuccessMessage()
	messageBody, err := json.Marshal(msgData)
	if err != nil {
		return fmt.Errorf("failed to marshal success message: %w", err)
	}

	messageID, err := uc.message.SendMessage(ctx, uc.outputQueueURL, string(messageBody))
	if err != nil {
		return fmt.Errorf("failed to send success message: %w", err)
	}

	log.Printf("[UseCase] Success message sent. MessageID: %s", messageID)
	return nil
}

func (uc *ProcessVideoUseCase) sendErrorMessage(ctx context.Context, result *domain.ProcessResult) error {
	log.Printf("[UseCase] Sending error message for process_id: %s. Error: %v", result.ProcessID, result.Error)

	msgData := result.ToErrorMessage()
	messageBody, err := json.Marshal(msgData)
	if err != nil {
		log.Printf("[UseCase] Failed to marshal error message: %v", err)
		return fmt.Errorf("failed to marshal error message: %w", err)
	}

	messageID, err := uc.message.SendMessage(ctx, uc.outputQueueURL, string(messageBody))
	if err != nil {
		log.Printf("[UseCase] Failed to send error message: %v", err)
		return fmt.Errorf("failed to send error message: %w", err)
	}

	log.Printf("[UseCase] Error message sent. MessageID: %s", messageID)
	return result.Error
}

func isValidVideoFile(filename string) bool {
	ext := filepath.Ext(filename)
	validExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}
