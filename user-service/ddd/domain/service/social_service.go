package service

import (
	"context"
	"fmt"

	"user-service/ddd/domain/entity"
	"user-service/ddd/domain/repo"
	"user-service/ddd/infrastructure/cache"
	"user-service/ddd/infrastructure/database/persistence"
	kafkainfra "user-service/ddd/infrastructure/kafka"
	"user-service/internal/resource"
)

// SocialService 封装关注/取关相关的领域逻辑：
// - 正常情况下：通过 Kafka 异步写 DB，同时立即在 Redis 中更新关注边，保证用户“是否关注”的查询是实时的；
// - 降级路径：当 Kafka 不可用时，退化为直接写数据库 + 缓存（FollowRepository 内部已处理）。
type SocialService struct {
	followRepo  repo.FollowRepository
	followCache *cache.FollowCache
}

// NewSocialService 使用默认的仓储和 Redis 客户端创建服务。
func NewSocialService() *SocialService {
	var followCache *cache.FollowCache
	if cli := resource.DefaultRedisResource().Client(); cli != nil {
		followCache = cache.NewFollowCache(cli)
	}
	return &SocialService{
		followRepo:  persistence.NewFollowRepository(),
		followCache: followCache,
	}
}

func (s *SocialService) markEdge(ctx context.Context, userUUID, targetUUID string, following bool) {
	if s == nil || s.followCache == nil {
		return
	}
	_ = s.followCache.SetEdge(ctx, userUUID, targetUUID, following)
}

// ToggleFollow 根据 action 统一处理关注/取关：
// - "follow"   -> Follow
// - "unfollow" -> Unfollow
func (s *SocialService) ToggleFollow(ctx context.Context, followEntity *entity.FollowEntity) error {
	if s == nil || followEntity == nil {
		return nil
	}
	userUUID := followEntity.UserUUID()
	targetUUID := followEntity.TargetUserUUID()
	status := followEntity.Status().Value()

	switch status {
	case kafkainfra.FollowOpFollow:
		// 优先通过 Kafka 异步落库，失败时退化为直接写 DB。
		if err := kafkainfra.PublishFollowEvent(ctx, kafkainfra.FollowOpFollow, userUUID, targetUUID); err != nil {
			return s.followRepo.Follow(ctx, userUUID, targetUUID)
		}
		// Kafka 成功后，先在缓存里标记为已关注，提升查询体验。
		s.markEdge(ctx, userUUID, targetUUID, true)
		return nil
	case kafkainfra.FollowOpUnfollow:
		if err := kafkainfra.PublishFollowEvent(ctx, kafkainfra.FollowOpUnfollow, userUUID, targetUUID); err != nil {
			return s.followRepo.Unfollow(ctx, userUUID, targetUUID)
		}
		s.markEdge(ctx, userUUID, targetUUID, false)
		return nil
	default:
		// 理论上不会走到这里，Action 已在 CQE 层校验。
		return fmt.Errorf("unsupported follow status: %s", status)
	}
}
