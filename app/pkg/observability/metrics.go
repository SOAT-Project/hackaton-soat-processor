package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ProcessedMessages tracks total messages processed by status
	ProcessedMessages = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_messages_processed_total",
			Help: "Total number of messages processed by the worker",
		},
		[]string{"status"},
	)

	// ProcessedVideos tracks total videos processed by status
	ProcessedVideos = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_videos_processed_total",
			Help: "Total number of videos processed by the worker",
		},
		[]string{"status"},
	)

	// ProcessingDuration tracks video processing duration
	ProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "worker_processing_duration_seconds",
			Help:    "Video processing duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600, 1200},
		},
		[]string{"status"},
	)

	// ExtractedFrames tracks frames extracted from last video
	ExtractedFrames = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_frames_extracted_last",
			Help: "Number of frames extracted from the last processed video",
		},
	)

	// ErrorsByType tracks errors by type
	ErrorsByType = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_errors_total",
			Help: "Total number of errors by type",
		},
		[]string{"type"},
	)

	// ActiveMessages tracks messages currently being processed
	ActiveMessages = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "worker_messages_active",
			Help: "Number of messages currently being processed",
		},
	)

	// FileSizes tracks file sizes in bytes
	FileSizes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "worker_file_size_bytes",
			Help:    "File sizes in bytes",
			Buckets: prometheus.ExponentialBuckets(1024*1024, 2, 10), // 1MB to 1GB
		},
		[]string{"type"},
	)

	// S3Operations tracks S3 operations
	S3Operations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_s3_operations_total",
			Help: "Total number of S3 operations",
		},
		[]string{"operation", "status"},
	)

	// SQSOperations tracks SQS operations
	SQSOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_sqs_operations_total",
			Help: "Total number of SQS operations",
		},
		[]string{"operation", "status"},
	)
)

// RecordMessageProcessed records a processed message
func RecordMessageProcessed(success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	ProcessedMessages.WithLabelValues(status).Inc()
}

// RecordVideoProcessed records a processed video with duration and frame count
func RecordVideoProcessed(success bool, duration float64, frames int) {
	status := "success"
	if !success {
		status = "error"
	}

	ProcessedVideos.WithLabelValues(status).Inc()
	ProcessingDuration.WithLabelValues(status).Observe(duration)

	if success && frames > 0 {
		ExtractedFrames.Set(float64(frames))
	}
}

// RecordError records an error by type
func RecordError(errorType string) {
	ErrorsByType.WithLabelValues(errorType).Inc()
}

// RecordS3Operation records an S3 operation
func RecordS3Operation(operation string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	S3Operations.WithLabelValues(operation, status).Inc()
}

// RecordSQSOperation records an SQS operation
func RecordSQSOperation(operation string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	SQSOperations.WithLabelValues(operation, status).Inc()
}

// RecordFileSize records a file size
func RecordFileSize(fileType string, size int64) {
	FileSizes.WithLabelValues(fileType).Observe(float64(size))
}

// IncrementActiveMessages increments active messages counter
func IncrementActiveMessages() {
	ActiveMessages.Inc()
}

// DecrementActiveMessages decrements active messages counter
func DecrementActiveMessages() {
	ActiveMessages.Dec()
}
