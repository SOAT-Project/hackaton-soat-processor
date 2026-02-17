package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SOAT-Project/hackaton-soat-processor/internal/adapter"
	"github.com/SOAT-Project/hackaton-soat-processor/internal/application/domain"
	"github.com/SOAT-Project/hackaton-soat-processor/internal/application/usecase"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/message"
	"github.com/SOAT-Project/hackaton-soat-processor/pkg/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var (
	inputQueueURL  = os.Getenv("QUEUE_INPUT")
	outputQueueURL = os.Getenv("QUEUE_OUTPUT")
	outputBucket   = os.Getenv("STORAGE_OUTPUT")
	region         = os.Getenv("AWS_REGION")
)

func main() {
	log.Println("🎬 Starting Video Processor Worker")

	// Valida variáveis de ambiente
	if err := validateEnvVars(); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}

	// Configura AWS
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Inicializa os serviços e adapters
	storageService := storage.NewS3Client(cfg)
	storagePort := adapter.NewStorageAdapter(storageService)

	messageService := message.NewSQSClient(cfg)
	messagePort := adapter.NewMessageAdapter(messageService)

	// Usa /tmp que sempre tem permissão de escrita para todos
	videoProcessor := adapter.NewFFmpegVideoProcessor("/tmp/video-processor")

	// Inicializa o use case
	processVideoUseCase := usecase.NewProcessVideoUseCase(
		storagePort,
		messagePort,
		videoProcessor,
		outputBucket,
		outputQueueURL,
	)

	// Inicializa o cliente SQS para consumo
	sqsClient := sqs.NewFromConfig(cfg)

	// Canal para graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Printf("✅ Worker initialized successfully")
	log.Printf("📥 Input Queue: %s", inputQueueURL)
	log.Printf("📤 Output Queue: %s", outputQueueURL)
	log.Printf("🪣 Output Bucket: %s", outputBucket)
	log.Println("⏳ Waiting for messages...")

	// Loop principal de processamento
	running := true
	for running {
		select {
		case <-sigChan:
			log.Println("🛑 Shutdown signal received, stopping worker...")
			running = false
			continue
		default:
		}

		// Recebe mensagens da fila
		res, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(inputQueueURL),
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     10,
			VisibilityTimeout:   300, // 5 minutos para processar
		})

		if err != nil {
			log.Printf("❌ Error receiving message: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Processa cada mensagem
		for _, msg := range res.Messages {
			if err := processMessage(ctx, processVideoUseCase, sqsClient, msg); err != nil {
				log.Printf("❌ Error processing message: %v", err)
			}
		}
	}

	log.Println("👋 Worker stopped gracefully")
}

func validateEnvVars() error {
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
		log.Printf("⚠️  AWS_REGION not set, using default: %s", region)
	}
	return nil
}

func processMessage(ctx context.Context, useCase *usecase.ProcessVideoUseCase, sqsClient *sqs.Client, msg types.Message) error {
	log.Printf("📨 Received message: ID=%s", *msg.MessageId)

	// Parse da mensagem
	var request struct {
		ProcessID   string `json:"process_id"`
		VideoBucket string `json:"video_bucket"`
		VideoKey    string `json:"video_key"`
	}

	if err := json.Unmarshal([]byte(*msg.Body), &request); err != nil {
		log.Printf("❌ Failed to parse message: %v", err)
		// Deleta mensagem inválida da fila
		deleteMessage(ctx, sqsClient, msg)
		return err
	}

	log.Printf("🎥 Processing video: process_id=%s, bucket=%s, key=%s",
		request.ProcessID, request.VideoBucket, request.VideoKey)

	// Cria o domínio
	videoProcess := domain.VideoProcess{
		ProcessID:   request.ProcessID,
		VideoBucket: request.VideoBucket,
		VideoKey:    request.VideoKey,
		CreatedAt:   time.Now(),
	}

	// Executa o use case
	startTime := time.Now()
	err := useCase.Execute(ctx, videoProcess)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ Processing failed for process_id=%s: %v (duration: %s)",
			request.ProcessID, err, duration)
	} else {
		log.Printf("✅ Processing completed for process_id=%s (duration: %s)",
			request.ProcessID, duration)
	}

	// Deleta mensagem da fila (tanto em sucesso quanto em erro, pois já enviamos a notificação)
	deleteMessage(ctx, sqsClient, msg)

	return err
}

func deleteMessage(ctx context.Context, sqsClient *sqs.Client, msg types.Message) {
	_, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(inputQueueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		log.Printf("⚠️  Failed to delete message from queue: %v", err)
	} else {
		log.Printf("🗑️  Message deleted from queue: ID=%s", *msg.MessageId)
	}
}
