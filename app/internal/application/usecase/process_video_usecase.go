package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SOAT-Project/hackaton-soat-processor/internal/application/domain"
	"github.com/SOAT-Project/hackaton-soat-processor/internal/port"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/observability"
	"go.uber.org/zap"
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
	startTime := time.Now()
	logger := observability.GetLogger().With(
		zap.String("process_id", request.ProcessID),
		zap.String("video_bucket", request.VideoBucket),
		zap.String("video_key", request.VideoKey),
	)

	observability.IncrementActiveMessages()
	defer observability.DecrementActiveMessages()

	logger.Info("starting video processing")

	result := &domain.ProcessResult{
		ProcessID: request.ProcessID,
		Success:   false,
	}

	if err := uc.validateRequest(request); err != nil {
		logger.Error("validation failed", zap.Error(err))
		observability.RecordError("validation")
		result.Error = err
		return uc.sendErrorMessage(ctx, result)
	}

	videoPath, err := uc.downloadVideo(ctx, request)
	if err != nil {
		logger.Error("video download failed", zap.Error(err))
		observability.RecordError("download")
		observability.RecordVideoProcessed(false, time.Since(startTime).Seconds(), 0)
		result.Error = fmt.Errorf("failed to download video: %w", err)
		return uc.sendErrorMessage(ctx, result)
	}
	defer os.Remove(videoPath)

	// Record video file size
	if stat, err := os.Stat(videoPath); err == nil {
		observability.RecordFileSize("video", stat.Size())
		logger.Info("video downloaded", zap.Int64("size_bytes", stat.Size()))
	}

	zipPath, frameCount, err := uc.videoProcessor.ProcessVideo(ctx, videoPath)
	if err != nil {
		logger.Error("video processing failed", zap.Error(err))
		observability.RecordError("processing")
		observability.RecordVideoProcessed(false, time.Since(startTime).Seconds(), 0)
		result.Error = fmt.Errorf("failed to process video: %w", err)
		return uc.sendErrorMessage(ctx, result)
	}
	defer os.Remove(zipPath)

	logger.Info("video processed successfully", zap.Int("frames_extracted", frameCount))

	// Record zip file size
	if stat, err := os.Stat(zipPath); err == nil {
		observability.RecordFileSize("zip", stat.Size())
		logger.Info("zip created", zap.Int64("size_bytes", stat.Size()))
	}

	outputKey := fmt.Sprintf("processed/frames_%s.zip", request.ProcessID)
	if err := uc.uploadZip(ctx, zipPath, outputKey); err != nil {
		logger.Error("zip upload failed", zap.Error(err))
		observability.RecordError("upload")
		observability.RecordVideoProcessed(false, time.Since(startTime).Seconds(), frameCount)
		result.Error = fmt.Errorf("failed to upload zip: %w", err)
		return uc.sendErrorMessage(ctx, result)
	}

	logger.Info("zip uploaded successfully", zap.String("output_key", outputKey))

	if err := uc.deleteOriginalVideo(ctx, request); err != nil {
		logger.Warn("failed to delete original video", zap.Error(err))
	} else {
		logger.Info("original video deleted successfully")
	}

	duration := time.Since(startTime)
	observability.RecordVideoProcessed(true, duration.Seconds(), frameCount)

	result.Success = true
	result.FileBucket = uc.outputBucket
	result.FileKey = outputKey

	logger.Info("video processing completed",
		zap.Duration("total_duration", duration),
		zap.Int("frames", frameCount),
	)

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
	logger := observability.GetLogger()
	logger.Info("downloading video from S3",
		zap.String("bucket", request.VideoBucket),
		zap.String("key", request.VideoKey),
	)

	body, err := uc.storage.GetObject(ctx, request.VideoBucket, request.VideoKey)
	if err != nil {
		observability.RecordS3Operation("get", false)
		return "", fmt.Errorf("failed to get object from storage: %w", err)
	}
	defer body.Close()

	observability.RecordS3Operation("get", true)

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

	logger.Debug("video downloaded successfully", zap.String("path", tempFile))
	return tempFile, nil
}

func (uc *ProcessVideoUseCase) uploadZip(ctx context.Context, zipPath, outputKey string) error {
	logger := observability.GetLogger()
	logger.Info("uploading ZIP to S3",
		zap.String("bucket", uc.outputBucket),
		zap.String("key", outputKey),
	)

	file, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer file.Close()

	_, err = uc.storage.PutObject(ctx, uc.outputBucket, outputKey, file)
	if err != nil {
		observability.RecordS3Operation("put", false)
		return fmt.Errorf("failed to put object to storage: %w", err)
	}

	observability.RecordS3Operation("put", true)
	return nil
}

func (uc *ProcessVideoUseCase) deleteOriginalVideo(ctx context.Context, request domain.VideoProcess) error {
	logger := observability.GetLogger()
	logger.Info("deleting original video from S3",
		zap.String("bucket", request.VideoBucket),
		zap.String("key", request.VideoKey),
	)

	err := uc.storage.DeleteObject(ctx, request.VideoBucket, request.VideoKey)
	if err != nil {
		observability.RecordS3Operation("delete", false)
		return fmt.Errorf("failed to delete original video: %w", err)
	}

	observability.RecordS3Operation("delete", true)
	return nil
}

func (uc *ProcessVideoUseCase) sendSuccessMessage(ctx context.Context, result *domain.ProcessResult) error {
	logger := observability.GetLogger()
	logger.Info("sending success message",
		zap.String("process_id", result.ProcessID),
		zap.String("file_key", result.FileKey),
	)

	msgData := result.ToSuccessMessage()
	messageBody, err := json.Marshal(msgData)
	if err != nil {
		return fmt.Errorf("failed to marshal success message: %w", err)
	}

	messageID, err := uc.message.SendMessage(ctx, uc.outputQueueURL, string(messageBody))
	if err != nil {
		observability.RecordSQSOperation("send", false)
		return fmt.Errorf("failed to send success message: %w", err)
	}

	observability.RecordSQSOperation("send", true)
	logger.Debug("success message sent", zap.String("message_id", messageID))
	return nil
}

func (uc *ProcessVideoUseCase) sendErrorMessage(ctx context.Context, result *domain.ProcessResult) error {
	logger := observability.GetLogger()
	logger.Error("sending error message",
		zap.String("process_id", result.ProcessID),
		zap.Error(result.Error),
	)

	msgData := result.ToErrorMessage()
	messageBody, err := json.Marshal(msgData)
	if err != nil {
		logger.Error("failed to marshal error message", zap.Error(err))
		return fmt.Errorf("failed to marshal error message: %w", err)
	}

	messageID, err := uc.message.SendMessage(ctx, uc.outputQueueURL, string(messageBody))
	if err != nil {
		observability.RecordSQSOperation("send", false)
		logger.Error("failed to send error message", zap.Error(err))
		return fmt.Errorf("failed to send error message: %w", err)
	}

	observability.RecordSQSOperation("send", true)
	logger.Debug("error message sent", zap.String("message_id", messageID))
	return result.Error
}

func isValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}
