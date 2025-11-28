package app

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"video-service/ddd/application/cqe"
	"video-service/ddd/application/dto"
	"video-service/ddd/domain/entity"
	"video-service/ddd/domain/service"
	"video-service/ddd/infrastructure/database/repository"
	"video-service/pkg/errno"
)

// VideoApp validates inputs and orchestrates DTOs.
type VideoApp interface {
	Publish(ctx context.Context, req *cqe.PublishVideoReq) (*dto.VideoDto, error)
	Precreate(ctx context.Context, req *cqe.PrecreateReq) (*dto.VideoDto, error)
	UpdateTranscodeResult(ctx context.Context, req *cqe.UpdateTranscodeResultReq) (*dto.VideoDto, error)
	Get(ctx context.Context, req *cqe.GetVideoReq) (*dto.VideoDto, error)
	List(ctx context.Context, req *cqe.ListVideosReq) (*dto.VideoListDto, error)
	Like(ctx context.Context, req *cqe.LikeReq) (*dto.LikeDto, error)
	Unlike(ctx context.Context, req *cqe.LikeReq) (*dto.LikeDto, error)
	Play(ctx context.Context, req *cqe.PlayReq) (*dto.PlayDto, error)
	AddComment(ctx context.Context, req *cqe.CommentCreateReq) (*dto.CommentDto, error)
	ListComments(ctx context.Context, req *cqe.ListCommentsReq) (*dto.CommentListDto, error)
}

type videoAppImpl struct {
	svc service.VideoService
}

// NewVideoApp builds the application layer using the provided DB handle.
func NewVideoApp(db *gorm.DB) VideoApp {
	videoRepo := repository.NewVideoRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	return &videoAppImpl{svc: service.NewVideoService(videoRepo, likeRepo, commentRepo)}
}

func DefaultVideoApp() VideoApp {
	videoRepo := repository.DefaultVideoRepository()
	likeRepo := repository.DefaultLikeRepository()
	commentRepo := repository.DefaultCommentRepository()
	return &videoAppImpl{svc: service.NewVideoService(videoRepo, likeRepo, commentRepo)}
}

func (v *videoAppImpl) Publish(ctx context.Context, req *cqe.PublishVideoReq) (*dto.VideoDto, error) {
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if req.VideoUUID == "" {
		req.VideoUUID = uuid.NewString()
	}
	video, err := v.svc.Publish(ctx, req)
	if err != nil {
		return nil, err
	}
	return dto.NewVideoDto(video), nil
}

func (v *videoAppImpl) Precreate(ctx context.Context, req *cqe.PrecreateReq) (*dto.VideoDto, error) {
	if req == nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "req")
	}
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if req.VideoUUID == "" {
		req.VideoUUID = uuid.NewString()
	}
	video, err := v.svc.Precreate(ctx, req)
	if err != nil {
		return nil, err
	}
	return dto.NewVideoDto(video), nil
}

func (v *videoAppImpl) UpdateTranscodeResult(ctx context.Context, req *cqe.UpdateTranscodeResultReq) (*dto.VideoDto, error) {
	if req == nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "req")
	}
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	video, err := v.svc.UpdateTranscodeResult(ctx, req)
	if err != nil {
		return nil, err
	}
	return dto.NewVideoDto(video), nil
}

func (v *videoAppImpl) Get(ctx context.Context, req *cqe.GetVideoReq) (*dto.VideoDto, error) {
	if req == nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "req")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	video, err := v.svc.Get(ctx, req.VideoUUID)
	if err != nil {
		return nil, err
	}
	return dto.NewVideoDto(video), nil
}

func (v *videoAppImpl) List(ctx context.Context, req *cqe.ListVideosReq) (*dto.VideoListDto, error) {
	if req == nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "req")
	}
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var videos []*entity.Video
	var total int64
	var err error
	if req.UserUUID != "" || req.Status != "" {
		videos, total, err = v.svc.ListByUserStatus(ctx, req.UserUUID, req.Status, req.Page, req.Size)
	} else {
		videos, total, err = v.svc.List(ctx, req.Page, req.Size)
	}
	if err != nil {
		return nil, err
	}
	return dto.NewVideoListDto(videos, total, req.Page, req.Size), nil
}

func (v *videoAppImpl) Like(ctx context.Context, req *cqe.LikeReq) (*dto.LikeDto, error) {
	if req == nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "req")
	}
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := v.svc.Like(ctx, req.UserUUID, req.VideoUUID); err != nil {
		return nil, err
	}
	video, err := v.svc.Get(ctx, req.VideoUUID)
	if err != nil {
		return &dto.LikeDto{VideoUUID: req.VideoUUID, UserUUID: req.UserUUID, Liked: true}, nil
	}
	return &dto.LikeDto{VideoUUID: req.VideoUUID, UserUUID: req.UserUUID, Liked: true, LikeCount: video.LikeCount}, nil
}

func (v *videoAppImpl) Unlike(ctx context.Context, req *cqe.LikeReq) (*dto.LikeDto, error) {
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := v.svc.Unlike(ctx, req.UserUUID, req.VideoUUID); err != nil {
		return nil, err
	}
	video, err := v.svc.Get(ctx, req.VideoUUID)
	if err != nil {
		return &dto.LikeDto{VideoUUID: req.VideoUUID, UserUUID: req.UserUUID, Liked: false}, nil
	}
	return &dto.LikeDto{VideoUUID: req.VideoUUID, UserUUID: req.UserUUID, Liked: false, LikeCount: video.LikeCount}, nil
}

func (v *videoAppImpl) Play(ctx context.Context, req *cqe.PlayReq) (*dto.PlayDto, error) {
	if req == nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "req")
	}
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := v.svc.Play(ctx, req.VideoUUID); err != nil {
		return nil, err
	}
	return &dto.PlayDto{VideoUUID: req.VideoUUID, Autoplay: true, Muted: false}, nil
}

func (v *videoAppImpl) AddComment(ctx context.Context, req *cqe.CommentCreateReq) (*dto.CommentDto, error) {
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	comment, err := v.svc.AddComment(ctx, req)
	if err != nil {
		return nil, err
	}
	return dto.NewCommentDto(comment), nil
}

func (v *videoAppImpl) ListComments(ctx context.Context, req *cqe.ListCommentsReq) (*dto.CommentListDto, error) {
	if req == nil {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "req")
	}
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, err
	}
	comments, total, err := v.svc.ListComments(ctx, req.VideoUUID, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	return dto.NewCommentListDto(comments, total, req.Page, req.Size), nil
}

// toVideoDto converts a domain video to DTO.
// converter functions moved to dto package
