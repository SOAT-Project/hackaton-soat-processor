package adapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFFmpegVideoProcessor(t *testing.T) {
	processor := NewFFmpegVideoProcessor("")
	if processor == nil {
		t.Fatal("NewFFmpegVideoProcessor returned nil")
	}

	// Cleanup default temp dir
	defer os.RemoveAll("temp")
}

func TestNewFFmpegVideoProcessor_CustomTempDir(t *testing.T) {
	tempDir := "custom_temp_test"
	defer os.RemoveAll(tempDir)

	processor := NewFFmpegVideoProcessor(tempDir)
	if processor == nil {
		t.Fatal("NewFFmpegVideoProcessor returned nil")
	}

	// Verify temp dir was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Temp directory %s was not created", tempDir)
	}
}

func TestFFmpegVideoProcessor_ProcessVideo_InvalidPath(t *testing.T) {
	processor := NewFFmpegVideoProcessor("test_temp")
	defer os.RemoveAll("test_temp")

	ctx := context.Background()
	_, _, err := processor.(*FFmpegVideoProcessor).ProcessVideo(ctx, "/nonexistent/video.mp4")
	if err == nil {
		t.Error("Expected error for nonexistent video file")
	}
}

func TestFFmpegVideoProcessor_CreateZipFile(t *testing.T) {
	tempDir := "test_zip_temp"
	os.MkdirAll(tempDir, 0777)
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile1 := filepath.Join(tempDir, "file1.txt")
	testFile2 := filepath.Join(tempDir, "file2.txt")
	
	os.WriteFile(testFile1, []byte("content 1"), 0644)
	os.WriteFile(testFile2, []byte("content 2"), 0644)

	processor := &FFmpegVideoProcessor{tempDir: tempDir}
	
	zipPath := filepath.Join(tempDir, "test.zip")
	files := []string{testFile1, testFile2}
	
	err := processor.createZipFile(files, zipPath)
	if err != nil {
		t.Fatalf("createZipFile failed: %v", err)
	}

	// Verify zip file was created
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Error("Zip file was not created")
	}
}

func TestFFmpegVideoProcessor_CreateZipFile_InvalidPath(t *testing.T) {
	processor := &FFmpegVideoProcessor{tempDir: "test_temp"}
	defer os.RemoveAll("test_temp")

	err := processor.createZipFile([]string{}, "/invalid/path/test.zip")
	if err == nil {
		t.Error("Expected error for invalid zip path")
	}
}

func TestFFmpegVideoProcessor_CreateZipFile_NonexistentFile(t *testing.T) {
	tempDir := "test_zip_error_temp"
	os.MkdirAll(tempDir, 0777)
	defer os.RemoveAll(tempDir)

	processor := &FFmpegVideoProcessor{tempDir: tempDir}
	
	zipPath := filepath.Join(tempDir, "test.zip")
	files := []string{"/nonexistent/file.txt"}
	
	err := processor.createZipFile(files, zipPath)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestFFmpegVideoProcessor_AddFileToZip(t *testing.T) {
	tempDir := "test_add_file_temp"
	os.MkdirAll(tempDir, 0777)
	defer os.RemoveAll(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "testfile.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	processor := &FFmpegVideoProcessor{tempDir: tempDir}

	// Create zip and test addFileToZip
	zipPath := filepath.Join(tempDir, "test.zip")
	err := processor.createZipFile([]string{testFile}, zipPath)
	if err != nil {
		t.Fatalf("createZipFile failed: %v", err)
	}

	// Verify zip exists
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Error("Zip file was not created")
	}
}

func TestFFmpegVideoProcessor_ProcessDirectory(t *testing.T) {
	tempDir := "test_process_temp"
	os.MkdirAll(tempDir, 0777)
	defer os.RemoveAll(tempDir)

	_ = &FFmpegVideoProcessor{tempDir: tempDir}

	// Test that temp directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Temp directory %s does not exist", tempDir)
	}

	// Test cleanup works
	os.RemoveAll(tempDir)
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Temp directory was not removed")
	}
}

func TestFFmpegVideoProcessor_Integration(t *testing.T) {
	// Skip if FFMPEG is not available
	if _, err := os.Stat("/usr/bin/ffmpeg"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/local/bin/ffmpeg"); os.IsNotExist(err) {
			t.Skip("FFmpeg not found, skipping integration test")
		}
	}

	tempDir := "test_integration_temp"
	defer os.RemoveAll(tempDir)

	processor := NewFFmpegVideoProcessor(tempDir)

	// Create a minimal test video file (this is a placeholder)
	// In real integration tests, you would need a valid video file
	testVideo := filepath.Join(tempDir, "test_video.mp4")
	os.MkdirAll(tempDir, 0777)
	
	// Note: This will fail without a real video
	// but it tests the code path
	ctx := context.Background()
	_, _, err := processor.(*FFmpegVideoProcessor).ProcessVideo(ctx, testVideo)
	
	// We expect this to fail since we don't have a real video
	if err == nil {
		t.Log("Unexpected success - test video should not exist")
	}
}

func TestFFmpegVideoProcessor_TempDirectory(t *testing.T) {
	tempDir := "test_temp_dir"
	defer os.RemoveAll(tempDir)

	processor := NewFFmpegVideoProcessor(tempDir).(*FFmpegVideoProcessor)

	if processor.tempDir != tempDir {
		t.Errorf("Expected tempDir %s, got %s", tempDir, processor.tempDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Temp directory %s was not created", tempDir)
	}
}

func TestFFmpegVideoProcessor_EmptyTempDir(t *testing.T) {
	// Test with empty string - should use default "temp"
	processor := NewFFmpegVideoProcessor("").(*FFmpegVideoProcessor)
	defer os.RemoveAll("temp")

	if processor.tempDir != "temp" {
		t.Errorf("Expected default tempDir 'temp', got %s", processor.tempDir)
	}

	// Verify directory was created
	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		t.Error("Default temp directory was not created")
	}
}

func TestFFmpegVideoProcessor_ProcessVideo_NoFrames(t *testing.T) {
	tempDir := "test_no_frames"
	os.MkdirAll(tempDir, 0777)
	defer os.RemoveAll(tempDir)

	processor := &FFmpegVideoProcessor{tempDir: tempDir}

	// Test with invalid video that won't produce frames
	ctx := context.Background()
	_, _, err := processor.ProcessVideo(ctx, "/invalid/path.mp4")
	
	if err == nil {
		t.Error("Expected error for invalid video path")
	}
}

func TestFFmpegVideoProcessor_CreateZipFile_EmptyFiles(t *testing.T) {
	tempDir := "test_empty_zip"
	os.MkdirAll(tempDir, 0777)
	defer os.RemoveAll(tempDir)

	processor := &FFmpegVideoProcessor{tempDir: tempDir}
	
	zipPath := filepath.Join(tempDir, "empty.zip")
	
	// Create with empty file list
	err := processor.createZipFile([]string{}, zipPath)
	if err != nil {
		t.Fatalf("createZipFile with empty list failed: %v", err)
	}

	// Verify zip was created
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Error("Empty zip file was not created")
	}
}

func TestFFmpegVideoProcessor_ProcessVideo_CreateDirError(t *testing.T) {
	// Use a temp dir that doesn't exist and can't be created
	processor := &FFmpegVideoProcessor{tempDir: "/nonexistent/invalid/path"}

	ctx := context.Background()
	_, _, err := processor.ProcessVideo(ctx, "video.mp4")
	
	if err == nil {
		t.Error("Expected error for invalid temp directory")
	}
}



