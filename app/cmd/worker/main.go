package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SOAT-Project/hackaton-soat-processor/internal/adapter"
	"github.com/SOAT-Project/hackaton-soat-processor/internal/application/domain"
	"github.com/SOAT-Project/hackaton-soat-processor/internal/application/usecase"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/message"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/observability"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

var (
	inputQueueURL  = os.Getenv("QUEUE_INPUT")
	outputQueueURL = os.Getenv("QUEUE_OUTPUT")
	outputBucket   = os.Getenv("STORAGE_OUTPUT")
	region         = os.Getenv("AWS_REGION")
)

func main() {
	// Initialize logger
	environment := getEnv("ENVIRONMENT", "development")
	if err := observability.InitLogger(environment); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer observability.Sync()

	logger := observability.GetLogger()
	logger.Info("starting video processor worker",
		zap.String("environment", environment),
		zap.String("version", "1.0.0"),
	)

	// Start metrics server
	metricsPort := 8080
	metricsServer := observability.NewMetricsServer(metricsPort)
	if err := metricsServer.Start(); err != nil {
		logger.Fatal("failed to start metrics server", zap.Error(err))
	}

	// Validate environment variables
	if err := validateEnvVars(); err != nil {
		logger.Fatal("environment validation failed", zap.Error(err))
	}

	logger.Info("configuration loaded",
		zap.String("input_queue", inputQueueURL),
		zap.String("output_queue", outputQueueURL),
		zap.String("output_bucket", outputBucket),
		zap.String("region", region),
		zap.Int("metrics_port", metricsPort),
	)

	// Configure AWS
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		logger.Fatal("failed to load AWS config", zap.Error(err))
	}

	// Initialize services and adapters
	storageService := storage.NewS3Client(cfg)
	storagePort := adapter.NewStorageAdapter(storageService)

	messageService := message.NewSQSClient(cfg)
	messagePort := adapter.NewMessageAdapter(messageService)

	// Use /tmp which always has write permission for all users
	videoProcessor := adapter.NewFFmpegVideoProcessor("/tmp/video-processor")

	// Initialize use case
	processVideoUseCase := usecase.NewProcessVideoUseCase(
		storagePort,
		messagePort,
		videoProcessor,
		outputBucket,
		outputQueueURL,
	)

	// Initialize SQS client for message consumption
	sqsClient := sqs.NewFromConfig(cfg)

	// Channel for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	logger.Info("worker initialized successfully")
	logger.Info("ready to process messages")

	// Main processing loop
	running := true
	for running {
		select {
		case <-sigChan:
			logger.Info("shutdown signal received, stopping worker")
			running = false
			continue
		default:
		}

		// Receive messages from queue
		res, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(inputQueueURL),
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     10,
			VisibilityTimeout:   300, // 5 minutos para processar
		})

		if err != nil {
			logger.Warn("error receiving message", zap.Error(err))
			observability.RecordSQSOperation("receive", false)
			time.Sleep(5 * time.Second)
			continue
		}

		observability.RecordSQSOperation("receive", true)

		// Process each message
		for _, msg := range res.Messages {
			if err := processMessage(ctx, processVideoUseCase, sqsClient, msg); err != nil {
				logger.Error("error processing message", zap.Error(err))
				observability.RecordMessageProcessed(false)
			} else {
				observability.RecordMessageProcessed(true)
			}
		}
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := metricsServer.Stop(shutdownCtx); err != nil {
		logger.Error("error stopping metrics server", zap.Error(err))
	}

	logger.Info("worker stopped gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func validateEnvVars() error {
	logger := observability.GetLogger()

	if inputQueueURL == "" {
		return fmt.Errorf("QUEUE_INPUT environment variable is required")
	}
	if outputQueueURL == "" {
		return fmt.Errorf("QUEUE_OUTPUT environment variable is required")
	}
	if outputBucket == "" {
		return fmt.Errorf("STORAGE_OUTPUT environment variable is required")
	}
	if region == "" {
		region = "us-east-1" // Default
		logger.Warn("AWS_REGION not set, using default", zap.String("region", region))
	}
	return nil
}

func processMessage(ctx context.Context, useCase *usecase.ProcessVideoUseCase, sqsClient *sqs.Client, msg types.Message) error {
	logger := observability.GetLogger().With(zap.String("message_id", *msg.MessageId))
	logger.Info("received message from queue")

	// Parse message
	var request struct {
		ProcessID   string `json:"process_id"`
		VideoBucket string `json:"video_bucket"`
		VideoKey    string `json:"video_key"`
	}

	if err := json.Unmarshal([]byte(*msg.Body), &request); err != nil {
		logger.Error("failed to parse message", zap.Error(err))
		// Delete invalid message from queue
		deleteMessage(ctx, sqsClient, msg)
		return err
	}

	logger.Info("message parsed successfully",
		zap.String("process_id", request.ProcessID),
		zap.String("video_bucket", request.VideoBucket),
		zap.String("video_key", request.VideoKey),
	)

	// Create domain object
	videoProcess := domain.VideoProcess{
		ProcessID:   request.ProcessID,
		VideoBucket: request.VideoBucket,
		VideoKey:    request.VideoKey,
		CreatedAt:   time.Now(),
	}

	// Execute use case
	err := useCase.Execute(ctx, videoProcess)

	// Delete message from queue (both on success and error, since we already sent notification)
	deleteMessage(ctx, sqsClient, msg)

	return err
}

func deleteMessage(ctx context.Context, sqsClient *sqs.Client, msg types.Message) {
	logger := observability.GetLogger()

	_, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(inputQueueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		logger.Warn("failed to delete message from queue",
			zap.String("message_id", *msg.MessageId),
			zap.Error(err),
		)
		observability.RecordSQSOperation("delete", false)
	} else {
		logger.Debug("message deleted from queue", zap.String("message_id", *msg.MessageId))
		observability.RecordSQSOperation("delete", true)
	}
}
