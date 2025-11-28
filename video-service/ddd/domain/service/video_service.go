package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-service/ddd/application/cqe"
	"video-service/ddd/domain/entity"
	"video-service/ddd/domain/repo"
	"video-service/ddd/domain/vo"
	"video-service/pkg/errno"
)

// VideoService 聚合视频领域的业务能力（视频、点赞、评论）。
type VideoService interface {
	Publish(ctx context.Context, req *cqe.PublishVideoReq) (*entity.Video, error)
	Precreate(ctx context.Context, req *cqe.PrecreateReq) (*entity.Video, error)
	UpdateTranscodeResult(ctx context.Context, req *cqe.UpdateTranscodeResultReq) (*entity.Video, error)
	Get(ctx context.Context, videoUUID string) (*entity.Video, error)
	List(ctx context.Context, page, size int) ([]*entity.Video, int64, error)
	ListByUserStatus(ctx context.Context, userUUID string, status string, page, size int) ([]*entity.Video, int64, error)
	Play(ctx context.Context, videoUUID string) error

	Like(ctx context.Context, userUUID, videoUUID string) error
	Unlike(ctx context.Context, userUUID, videoUUID string) error

	AddComment(ctx context.Context, req *cqe.CommentCreateReq) (*entity.Comment, error)
	ListComments(ctx context.Context, videoUUID string, page, size int) ([]*entity.Comment, int64, error)
}

type videoServiceImpl struct {
	videoRepo   repo.VideoRepository
	likeRepo    repo.LikeRepository
	commentRepo repo.CommentRepository
}

func NewVideoService(videoRepo repo.VideoRepository, likeRepo repo.LikeRepository, commentRepo repo.CommentRepository) VideoService {
	return &videoServiceImpl{
		videoRepo:   videoRepo,
		likeRepo:    likeRepo,
		commentRepo: commentRepo,
	}
}

func (s *videoServiceImpl) Publish(ctx context.Context, req *cqe.PublishVideoReq) (*entity.Video, error) {
	video := &entity.Video{
		VideoUUID:         req.VideoUUID,
		UserUUID:          req.UserUUID,
		UploadVideoUUID:   req.UploadVideoUUID,
		Title:             req.Title,
		Description:       req.Description,
		CoverURL:          req.CoverURL,
		VideoURL:          req.VideoURL,
		DurationSec:       req.DurationSec,
		Resolution:        req.Resolution,
		SizeBytes:         req.SizeBytes,
		Status:            vo.NewVideoStatus(strings.ToLower(req.Status)).Value(),
		Privacy:           vo.VideoPrivacyPublic.Value(),
		TranscodeTaskUUID: "",
	}
	if video.Status == "" {
		video.Status = "processing"
	}
	now := time.Now()
	video.CreatedAt = now
	video.UpdatedAt = now
	if video.Status == "published" {
		video.PublishedAt = &now
	}
	if err := s.videoRepo.Create(ctx, video); err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	return video, nil
}

func (s *videoServiceImpl) Precreate(ctx context.Context, req *cqe.PrecreateReq) (*entity.Video, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	existed, err := s.videoRepo.FindByUUID(ctx, req.VideoUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	if existed != nil {
		return existed, nil
	}
	now := time.Now()
	video := &entity.Video{
		VideoUUID:         req.VideoUUID,
		UserUUID:          req.UserUUID,
		UploadVideoUUID:   req.UploadVideoUUID,
		Title:             req.Title,
		Description:       req.Description,
		CoverURL:          req.CoverURL,
		Status:            vo.VideoStatusProcessing.Value(),
		Privacy:           vo.VideoPrivacyPublic.Value(),
		TranscodeTaskUUID: req.TranscodeTaskUUID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.videoRepo.Create(ctx, video); err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	return video, nil
}

func (s *videoServiceImpl) UpdateTranscodeResult(ctx context.Context, req *cqe.UpdateTranscodeResultReq) (*entity.Video, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	video, err := s.videoRepo.FindByUUID(ctx, req.VideoUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return nil, errno.ErrNotFound
	}
	// 幂等与顺序：仅当任务匹配或未设置任务时允许更新；不接受过时回滚
	if video.TranscodeTaskUUID != "" && video.TranscodeTaskUUID != req.TaskUUID {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "task_uuid mismatch")
	}
	cur := vo.NewVideoStatus(strings.ToLower(video.Status)).Value()
	next := vo.NewVideoStatus(strings.ToLower(req.Status)).Value()
	if cur == "published" || cur == "failed" {
		// 已终态，忽略回到 processing 的请求
		if next == "processing" {
			return video, nil
		}
	}
	// 应用更新
	video.TranscodeTaskUUID = req.TaskUUID
	video.Status = next
	if next == "published" {
		video.VideoURL = req.VideoURL
		now := time.Now()
		video.PublishedAt = &now
	} else if next == "failed" {
		video.ErrorMessage = req.ErrorMsg
	}
	if req.DurationSec != nil {
		video.DurationSec = req.DurationSec
	}
	if req.SizeBytes != nil {
		video.SizeBytes = req.SizeBytes
	}
	video.UpdatedAt = time.Now()
	if err := s.videoRepo.Update(ctx, video); err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	return video, nil
}

func (s *videoServiceImpl) Get(ctx context.Context, videoUUID string) (*entity.Video, error) {
	video, err := s.videoRepo.FindByUUID(ctx, videoUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return nil, errno.ErrNotFound
	}
	s.fillCounts(ctx, video)
	return video, nil
}

func (s *videoServiceImpl) List(ctx context.Context, page, size int) ([]*entity.Video, int64, error) {
	videos, total, err := s.videoRepo.List(ctx, page, size)
	if err != nil {
		return nil, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	for _, v := range videos {
		s.fillCounts(ctx, v)
	}
	return videos, total, nil
}

func (s *videoServiceImpl) ListByUserStatus(ctx context.Context, userUUID string, status string, page, size int) ([]*entity.Video, int64, error) {
	videos, total, err := s.videoRepo.ListByUserStatus(ctx, userUUID, status, page, size)
	if err != nil {
		return nil, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	for _, v := range videos {
		s.fillCounts(ctx, v)
	}
	return videos, total, nil
}

func (s *videoServiceImpl) Play(ctx context.Context, videoUUID string) error {
	// TODO: persist play count; for now ensure video exists.
	video, err := s.videoRepo.FindByUUID(ctx, videoUUID)
	if err != nil {
		return errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return errno.ErrNotFound
	}
	return nil
}

func (s *videoServiceImpl) fillCounts(ctx context.Context, video *entity.Video) {
	if video == nil {
		return
	}
	var likeCount, commentCount int64
	if s.likeRepo != nil {
		if cnt, err := s.likeRepo.CountByVideo(ctx, video.VideoUUID); err == nil {
			likeCount = cnt
		}
	}
	if s.commentRepo != nil {
		if cnt, err := s.commentRepo.CountByVideo(ctx, video.VideoUUID); err == nil {
			commentCount = cnt
		}
	}
	video.SetCounts(likeCount, 0, commentCount)
}
func (s *videoServiceImpl) Like(ctx context.Context, userUUID, videoUUID string) error {
	video, err := s.videoRepo.FindByUUID(ctx, videoUUID)
	if err != nil {
		return errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return errno.ErrNotFound
	}
	_, err = s.likeRepo.Add(ctx, videoUUID, userUUID)
	if err != nil {
		return errno.NewBizError(errno.ErrDatabase, err)
	}
	return nil
}

func (s *videoServiceImpl) Unlike(ctx context.Context, userUUID, videoUUID string) error {
	video, err := s.videoRepo.FindByUUID(ctx, videoUUID)
	if err != nil {
		return errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return errno.ErrNotFound
	}
	if err := s.likeRepo.Remove(ctx, videoUUID, userUUID); err != nil {
		return errno.NewBizError(errno.ErrDatabase, err)
	}
	return nil
}

func (s *videoServiceImpl) AddComment(ctx context.Context, req *cqe.CommentCreateReq) (*entity.Comment, error) {
	video, err := s.videoRepo.FindByUUID(ctx, req.VideoUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return nil, errno.ErrNotFound
	}
	comment := &entity.Comment{
		CommentUUID: uuid.NewString(),
		VideoUUID:   req.VideoUUID,
		UserUUID:    req.UserUUID,
		Content:     req.Content,
		ParentUUID:  req.ParentUUID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	return comment, nil
}

func (s *videoServiceImpl) ListComments(ctx context.Context, videoUUID string, page, size int) ([]*entity.Comment, int64, error) {
	video, err := s.videoRepo.FindByUUID(ctx, videoUUID)
	if err != nil {
		return nil, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return nil, 0, errno.ErrNotFound
	}
	comments, total, err := s.commentRepo.ListByVideo(ctx, videoUUID, page, size)
	if err != nil {
		return nil, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	return comments, total, nil
}
