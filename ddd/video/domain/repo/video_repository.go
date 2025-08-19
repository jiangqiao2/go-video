package repo

import (
	"context"
	"go-video/ddd/video/domain/entity"
)

// VideoRepository 视频仓储接口
type VideoRepository interface {
	// Save 保存视频
	Save(ctx context.Context, video *entity.Video) error
}
