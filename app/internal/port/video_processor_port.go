package port

import "context"

type VideoProcessorPort interface {
	ProcessVideo(ctx context.Context, videoPath string) (zipPath string, frameCount int, err error)
}
