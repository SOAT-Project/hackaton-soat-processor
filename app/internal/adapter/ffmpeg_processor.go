package adapter

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/SOAT-Project/hackaton-soat-processor/internal/port"
)

type FFmpegVideoProcessor struct {
	tempDir string
}

func NewFFmpegVideoProcessor(tempDir string) port.VideoProcessorPort {
	if tempDir == "" {
		tempDir = "temp"
	}

	os.MkdirAll(tempDir, 0777)
	return &FFmpegVideoProcessor{
		tempDir: tempDir,
	}
}

func (p *FFmpegVideoProcessor) ProcessVideo(ctx context.Context, videoPath string) (string, int, error) {
	processDir := filepath.Join(p.tempDir, fmt.Sprintf("process_%d", os.Getpid()))
	if err := os.MkdirAll(processDir, 0777); err != nil {
		return "", 0, fmt.Errorf("failed to create process directory: %w", err)
	}
	defer os.RemoveAll(processDir)

	framePattern := filepath.Join(processDir, "frame_%04d.png")
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", videoPath,
		"-vf", "fps=1",
		"-y",
		framePattern,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", 0, fmt.Errorf("ffmpeg error: %w, output: %s", err, string(output))
	}

	frames, err := filepath.Glob(filepath.Join(processDir, "*.png"))
	if err != nil {
		return "", 0, fmt.Errorf("failed to list video frames: %w", err)
	}

	if len(frames) == 0 {
		return "", 0, fmt.Errorf("no frames extracted from video")
	}

	zipPath := filepath.Join(p.tempDir, fmt.Sprintf("frames_%d.zip", os.Getpid()))
	if err := p.createZipFile(frames, zipPath); err != nil {
		return "", 0, fmt.Errorf("failed to create zip: %w", err)
	}

	return zipPath, len(frames), nil
}

func (p *FFmpegVideoProcessor) createZipFile(files []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err := p.addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}

	return nil
}

func (p *FFmpegVideoProcessor) addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
