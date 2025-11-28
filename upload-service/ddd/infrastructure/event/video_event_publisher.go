package event

import (
	"context"
	"fmt"
	"sync"

	"upload-service/ddd/application/dto"
	"upload-service/ddd/domain/entity"
	"upload-service/internal/resource"
	"upload-service/pkg/logger"
	"upload-service/pkg/sse"
)

const (
	videoStatusEventName = "video.status.changed"
	videoTopicTemplate   = "user:%s"
)

// VideoEventPublisher publishes SSE notifications for video state changes.
type VideoEventPublisher interface {
	PublishStatusChanged(ctx context.Context, video *entity.VideoEntity) error
}

type redisVideoEventPublisher struct {
	publisher *sse.Publisher
}

type noopVideoEventPublisher struct{}

func (noopVideoEventPublisher) PublishStatusChanged(_ context.Context, _ *entity.VideoEntity) error {
	return nil
}

var (
	videoPublisher     VideoEventPublisher = noopVideoEventPublisher{}
	videoPublisherOnce sync.Once
)

// DefaultVideoEventPublisher returns a singleton publisher backed by the shared Redis resource.
func DefaultVideoEventPublisher() VideoEventPublisher {
	videoPublisherOnce.Do(func() {
		client := resource.DefaultRedisResource().Client()
		if client == nil {
			logger.Warnf("Redis resource unavailable, SSE video notifications disabled")
			return
		}
		videoPublisher = &redisVideoEventPublisher{
			publisher: sse.NewPublisher(client),
		}
	})
	return videoPublisher
}

// NewVideoEventPublisher constructs a publisher from an explicit SSE publisher instance.
func NewVideoEventPublisher(publisher *sse.Publisher) VideoEventPublisher {
	if publisher == nil {
		return noopVideoEventPublisher{}
	}
	return &redisVideoEventPublisher{publisher: publisher}
}

func (p *redisVideoEventPublisher) PublishStatusChanged(ctx context.Context, video *entity.VideoEntity) error {
	if p.publisher == nil {
		return fmt.Errorf("sse publisher not configured")
	}
	if video == nil {
		return fmt.Errorf("video entity is nil")
	}

	videoDTO := dto.NewVideoDetailDto(video)
	if videoDTO == nil {
		return fmt.Errorf("failed to convert video dto")
	}

	topic := fmt.Sprintf(videoTopicTemplate, video.UserUUID())
	if _, err := p.publisher.Publish(ctx, topic, video.UserUUID(), videoStatusEventName, videoDTO); err != nil {
		return err
	}
	return nil
}
