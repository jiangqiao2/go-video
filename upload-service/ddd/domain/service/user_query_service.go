package service

import (
    "context"
    "upload-service/ddd/domain/entity"
    grpcClient "upload-service/ddd/infrastructure/grpc"
    "upload-service/pkg/logger"
)

type UserSummary struct {
    UserUUID  string
    Account   string
    AvatarUrl string
}

type UserQueryService interface {
    GetUsersForVideos(ctx context.Context, videos []*entity.VideoEntity) (map[string]*UserSummary, error)
    BatchQueryUsers(ctx context.Context, userUUIDs []string) (map[string]*UserSummary, error)
}

type userQueryServiceImpl struct {
    client *grpcClient.UserServiceClient
}

func NewUserQueryService() UserQueryService {
    return &userQueryServiceImpl{
        client: grpcClient.DefaultUserServiceClient(),
    }
}

func (s *userQueryServiceImpl) GetUsersForVideos(ctx context.Context, videos []*entity.VideoEntity) (map[string]*UserSummary, error) {
    uuidSet := make(map[string]struct{}, len(videos))
    uuids := make([]string, 0, len(videos))
    for _, v := range videos {
        u := v.UserUUID()
        if u == "" {
            continue
        }
        if _, ok := uuidSet[u]; !ok {
            uuidSet[u] = struct{}{}
            uuids = append(uuids, u)
        }
    }
    return s.BatchQueryUsers(ctx, uuids)
}

func (s *userQueryServiceImpl) BatchQueryUsers(ctx context.Context, userUUIDs []string) (map[string]*UserSummary, error) {
    res := make(map[string]*UserSummary, len(userUUIDs))
    if len(userUUIDs) == 0 {
        return res, nil
    }
    infos, err := s.client.GetUsersByUUIDs(ctx, userUUIDs)
    if err != nil {
        logger.Warnf("GetUsersByUUIDs failed: %v", err)
        return res, nil
    }
    for _, info := range infos {
        if info == nil {
            continue
        }
        res[info.UserUuid] = &UserSummary{
            UserUUID:  info.UserUuid,
            Account:   info.Account,
            AvatarUrl: normalizeAvatarURL(info.AvatarUrl),
        }
    }
    return res, nil
}

func normalizeAvatarURL(url string) string {
    if url == "" {
        return ""
    }
    if len(url) >= 4 && (url[0:4] == "http" || url[0:5] == "https") {
        return url
    }
    return "http://localhost:9000/image/" + url
}