package domain

import (
	"errors"
	"testing"
	"time"
)

func TestVideoProcess_Creation(t *testing.T) {
	now := time.Now()
	vp := VideoProcess{
		ProcessID:   "test-123",
		VideoBucket: "test-bucket",
		VideoKey:    "video.mp4",
		CreatedAt:   now,
	}

	if vp.ProcessID != "test-123" {
		t.Errorf("Expected ProcessID test-123, got %s", vp.ProcessID)
	}
	if vp.VideoBucket != "test-bucket" {
		t.Errorf("Expected VideoBucket test-bucket, got %s", vp.VideoBucket)
	}
	if vp.VideoKey != "video.mp4" {
		t.Errorf("Expected VideoKey video.mp4, got %s", vp.VideoKey)
	}
	if vp.CreatedAt != now {
		t.Errorf("Expected CreatedAt %v, got %v", now, vp.CreatedAt)
	}
}

func TestProcessResult_ToSuccessMessage(t *testing.T) {
	result := ProcessResult{
		ProcessID:  "process-123",
		FileBucket: "output-bucket",
		FileKey:    "frames.zip",
		Success:    true,
		Error:      nil,
	}

	msg := result.ToSuccessMessage()

	if msg["process_id"] != "process-123" {
		t.Errorf("Expected process_id process-123, got %v", msg["process_id"])
	}
	if msg["file_bucket"] != "output-bucket" {
		t.Errorf("Expected file_bucket output-bucket, got %v", msg["file_bucket"])
	}
	if msg["file_key"] != "frames.zip" {
		t.Errorf("Expected file_key frames.zip, got %v", msg["file_key"])
	}
}

func TestProcessResult_ToErrorMessage_WithError(t *testing.T) {
	testError := errors.New("processing failed")
	result := ProcessResult{
		ProcessID: "process-456",
		Success:   false,
		Error:     testError,
	}

	msg := result.ToErrorMessage()

	if msg["process_id"] != "process-456" {
		t.Errorf("Expected process_id process-456, got %v", msg["process_id"])
	}
	if msg["error_message"] != "processing failed" {
		t.Errorf("Expected error_message 'processing failed', got %v", msg["error_message"])
	}
}

func TestProcessResult_ToErrorMessage_WithNilError(t *testing.T) {
	result := ProcessResult{
		ProcessID: "process-789",
		Success:   false,
		Error:     nil,
	}

	msg := result.ToErrorMessage()

	if msg["process_id"] != "process-789" {
		t.Errorf("Expected process_id process-789, got %v", msg["process_id"])
	}
	if msg["error_message"] != "unknown error" {
		t.Errorf("Expected error_message 'unknown error', got %v", msg["error_message"])
	}
}

func TestProcessResult_Creation(t *testing.T) {
	result := ProcessResult{
		ProcessID:  "test-id",
		FileBucket: "test-bucket",
		FileKey:    "test-key",
		Success:    true,
		Error:      nil,
	}

	if result.ProcessID != "test-id" {
		t.Errorf("Expected ProcessID test-id, got %s", result.ProcessID)
	}
	if result.FileBucket != "test-bucket" {
		t.Errorf("Expected FileBucket test-bucket, got %s", result.FileBucket)
	}
	if result.FileKey != "test-key" {
		t.Errorf("Expected FileKey test-key, got %s", result.FileKey)
	}
	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", result.Error)
	}
}

func TestProcessResult_SuccessAndErrorMessages(t *testing.T) {
	tests := []struct {
		name           string
		result         ProcessResult
		expectedFields map[string]string
		isSuccess      bool
	}{
		{
			name: "success with all fields",
			result: ProcessResult{
				ProcessID:  "success-1",
				FileBucket: "bucket-1",
				FileKey:    "key-1",
				Success:    true,
			},
			expectedFields: map[string]string{
				"process_id":  "success-1",
				"file_bucket": "bucket-1",
				"file_key":    "key-1",
			},
			isSuccess: true,
		},
		{
			name: "error with custom message",
			result: ProcessResult{
				ProcessID: "error-1",
				Success:   false,
				Error:     errors.New("custom error"),
			},
			expectedFields: map[string]string{
				"process_id":    "error-1",
				"error_message": "custom error",
			},
			isSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg map[string]interface{}
			if tt.isSuccess {
				msg = tt.result.ToSuccessMessage()
			} else {
				msg = tt.result.ToErrorMessage()
			}

			for key, expectedValue := range tt.expectedFields {
				if msg[key] != expectedValue {
					t.Errorf("Expected %s=%s, got %v", key, expectedValue, msg[key])
				}
			}
		})
	}
}
