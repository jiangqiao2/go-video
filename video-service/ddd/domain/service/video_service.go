package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-service/ddd/application/cqe"
	"video-service/ddd/domain/entity"
	"video-service/ddd/domain/gateway"
	"video-service/ddd/domain/repo"
	"video-service/ddd/domain/vo"
	"video-service/pkg/errno"
	"video-service/pkg/logger"
)

// VideoService 聚合视频领域的业务能力（视频、点赞、评论）。
type VideoService interface {
	Publish(ctx context.Context, req *cqe.PublishVideoReq) (*entity.Video, error)
	Precreate(ctx context.Context, req *cqe.PrecreateReq) (*entity.Video, error)
	UpdateTranscodeResult(ctx context.Context, req *cqe.UpdateTranscodeResultReq) (*entity.Video, error)
	Get(ctx context.Context, videoUUID string, userUUID string) (*entity.Video, bool, error)
	List(ctx context.Context, page, size int) ([]*entity.Video, int64, error)
	ListByUserStatus(ctx context.Context, userUUID string, status string, page, size int) ([]*entity.Video, int64, error)
	Play(ctx context.Context, videoUUID string) error

	ToggleLike(ctx context.Context, userUUID, videoUUID string) (bool, int64, error)

	AddComment(ctx context.Context, req *cqe.CommentCreateReq) (*entity.Comment, error)
	ListComments(ctx context.Context, videoUUID, rootUUID, parentUUID, sortBy string, page, size int, userUUID string) ([]*entity.Comment, int64, error)
	ToggleCommentLike(ctx context.Context, userUUID, commentUUID string) (bool, int64, error)
}

type videoServiceImpl struct {
	videoRepo       repo.VideoRepository
	likeRepo        repo.LikeRepository
	commentRepo     repo.CommentRepository
	commentLikeRepo repo.CommentLikeRepository
	notificationSvc gateway.NotificationService
}

func NewVideoService(videoRepo repo.VideoRepository, likeRepo repo.LikeRepository, commentRepo repo.CommentRepository, commentLikeRepo repo.CommentLikeRepository, notificationSvc gateway.NotificationService) VideoService {
	return &videoServiceImpl{
		videoRepo:       videoRepo,
		likeRepo:        likeRepo,
		commentRepo:     commentRepo,
		commentLikeRepo: commentLikeRepo,
		notificationSvc: notificationSvc,
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
	if video.Status == "published" {
		s.notifyVideoPublished(ctx, video)
	}
	return video, nil
}

func (s *videoServiceImpl) Precreate(ctx context.Context, req *cqe.PrecreateReq) (*entity.Video, error) {
	existed, err := s.videoRepo.FindByUUID(ctx, req.VideoUUID)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return video, nil
}

func (s *videoServiceImpl) UpdateTranscodeResult(ctx context.Context, req *cqe.UpdateTranscodeResultReq) (*entity.Video, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// 记录入口日志
	logger.Infof("UpdateTranscodeResult begin video_uuid=%s task_uuid=%s status=%s url=%s duration=%v size=%v", req.VideoUUID, req.TaskUUID, req.Status, req.VideoURL, req.DurationSec, req.SizeBytes)
	video, err := s.videoRepo.FindByUUID(ctx, req.VideoUUID)
	if err != nil {
		logger.Warnf("UpdateTranscodeResult find video failed video_uuid=%s error=%v", req.VideoUUID, err)
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		logger.Warnf("UpdateTranscodeResult video not found video_uuid=%s", req.VideoUUID)
		return nil, errno.ErrNotFound
	}
	// 幂等与顺序：仅当任务匹配或未设置任务时允许更新；不接受过时回滚
	if video.TranscodeTaskUUID != "" && video.TranscodeTaskUUID != req.TaskUUID {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "task_uuid mismatch")
	}
	cur := vo.NewVideoStatus(strings.ToLower(video.Status)).Value()
	next := vo.NewVideoStatus(strings.ToLower(req.Status)).Value()
	publishedJustNow := cur != "published" && next == "published"
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
	logger.Infof("UpdateTranscodeResult apply update video_uuid=%s task_uuid=%s status=%s url=%s", video.VideoUUID, video.TranscodeTaskUUID, video.Status, video.VideoURL)
	if err := s.videoRepo.Update(ctx, video); err != nil {
		logger.Warnf("UpdateTranscodeResult update failed video_uuid=%s error=%v", video.VideoUUID, err)
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	logger.Infof("UpdateTranscodeResult success video_uuid=%s", video.VideoUUID)
	if publishedJustNow {
		s.notifyVideoPublished(ctx, video)
	}
	return video, nil
}

func (s *videoServiceImpl) Get(ctx context.Context, videoUUID string, userUUID string) (*entity.Video, bool, error) {
	video, err := s.videoRepo.FindByUUID(ctx, videoUUID)
	if err != nil {
		return nil, false, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return nil, false, errno.ErrNotFound
	}
	s.fillCounts(ctx, video)
	liked := false
	if userUUID != "" && s.likeRepo != nil {
		if ok, err := s.likeRepo.Exists(ctx, video.VideoUUID, userUUID); err == nil {
			liked = ok
		}
	}
	return video, liked, nil
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
		if cnt, err := s.commentRepo.CountRootsByVideo(ctx, video.VideoUUID); err == nil {
			commentCount = cnt
		}
	}
	video.SetCounts(likeCount, 0, commentCount)
}

// notifyVideoPublished 在视频成功发布后向作者发送站内通知（失败不影响主流程）。
func (s *videoServiceImpl) notifyVideoPublished(ctx context.Context, video *entity.Video) {
	if s.notificationSvc == nil || video == nil {
		return
	}
	if video.UserUUID == "" || video.VideoUUID == "" {
		return
	}
	if err := s.notificationSvc.NotifyVideoPublished(ctx, video.UserUUID, video.VideoUUID, video.Title); err != nil {
		logger.Warnf("notify video published failed video_uuid=%s user_uuid=%s err=%v", video.VideoUUID, video.UserUUID, err)
	}
}

func (s *videoServiceImpl) ToggleLike(ctx context.Context, userUUID, videoUUID string) (bool, int64, error) {
	video, err := s.videoRepo.FindByUUID(ctx, videoUUID)
	if err != nil {
		return false, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return false, 0, errno.ErrNotFound
	}
	liked := false
	if s.likeRepo != nil {
		if ok, err := s.likeRepo.Exists(ctx, videoUUID, userUUID); err == nil && ok {
			liked = true
		}
	}
	var likeCount int64
	if liked {
		if err := s.likeRepo.Remove(ctx, videoUUID, userUUID); err != nil {
			return false, 0, errno.NewBizError(errno.ErrDatabase, err)
		}
	} else {
		if _, err := s.likeRepo.Add(ctx, videoUUID, userUUID); err != nil {
			return false, 0, errno.NewBizError(errno.ErrDatabase, err)
		}
	}
	if cnt, err := s.likeRepo.CountByVideo(ctx, videoUUID); err == nil {
		likeCount = cnt
	}
	return !liked, likeCount, nil
}

func (s *videoServiceImpl) AddComment(ctx context.Context, req *cqe.CommentCreateReq) (*entity.Comment, error) {
	video, err := s.videoRepo.FindByUUID(ctx, req.VideoUUID)
	if err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return nil, errno.ErrNotFound
	}
	now := time.Now()
	if req.ParentUUID == "" {
		root := &entity.Comment{
			CommentUUID: uuid.NewString(),
			RootUUID:    "",
			VideoUUID:   req.VideoUUID,
			UserUUID:    req.UserUUID,
			Content:     req.Content,
			LikeCount:   0,
			ReplyCount:  0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		root.RootUUID = root.CommentUUID
		if err := s.commentRepo.CreateRoot(ctx, root); err != nil {
			return nil, errno.NewBizError(errno.ErrDatabase, err)
		}
		return root, nil
	}

	// reply
	// find parent (root or reply)
	var parent *entity.Comment
	if p, err := s.commentRepo.FindRootByUUID(ctx, req.ParentUUID); err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	} else if p != nil {
		parent = p
	} else if p2, err := s.commentRepo.FindReplyByUUID(ctx, req.ParentUUID); err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	} else {
		parent = p2
	}
	if parent == nil || parent.VideoUUID != req.VideoUUID {
		return nil, errno.NewSimpleBizError(errno.ErrParameterInvalid, nil, "parent comment not found")
	}

	reply := &entity.Comment{
		CommentUUID: uuid.NewString(),
		RootUUID:    parent.RootUUID,
		VideoUUID:   req.VideoUUID,
		UserUUID:    req.UserUUID,
		Content:     req.Content,
		ParentUUID:  req.ParentUUID,
		ParentType:  "root",
		Depth:       parent.Depth + 1,
		Path:        parent.Path,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if parent.ParentUUID != "" {
		reply.ParentType = "reply"
	}
	if reply.RootUUID == "" {
		reply.RootUUID = parent.RootUUID
	}
	if err := s.commentRepo.CreateReply(ctx, reply); err != nil {
		return nil, errno.NewBizError(errno.ErrDatabase, err)
	}
	// bump counts
	if parent.ParentUUID == "" {
		_ = s.commentRepo.IncrementRootReplyCount(ctx, parent.CommentUUID, 1)
	} else {
		_ = s.commentRepo.IncrementReplyCount(ctx, parent.CommentUUID, 1)
		_ = s.commentRepo.IncrementRootReplyCount(ctx, parent.RootUUID, 1)
	}
	return reply, nil
}

func (s *videoServiceImpl) ListComments(ctx context.Context, videoUUID, rootUUID, parentUUID, sortBy string, page, size int, userUUID string) ([]*entity.Comment, int64, error) {
	video, err := s.videoRepo.FindByUUID(ctx, videoUUID)
	if err != nil {
		return nil, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	if video == nil {
		return nil, 0, errno.ErrNotFound
	}
	if rootUUID == "" {
		comments, total, err := s.commentRepo.ListRootsByVideo(ctx, videoUUID, sortBy, page, size)
		if err != nil {
			return nil, 0, errno.NewBizError(errno.ErrDatabase, err)
		}
		if userUUID != "" && s.commentLikeRepo != nil {
			for _, c := range comments {
				if ok, err := s.commentLikeRepo.Exists(ctx, c.CommentUUID, userUUID); err == nil && ok {
					c.Liked = true
				}
			}
		}
		return comments, total, nil
	}

	comments, total, err := s.commentRepo.ListReplies(ctx, rootUUID, parentUUID, sortBy, page, size)
	if err != nil {
		return nil, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	if userUUID != "" && s.commentLikeRepo != nil {
		for _, c := range comments {
			if ok, err := s.commentLikeRepo.Exists(ctx, c.CommentUUID, userUUID); err == nil && ok {
				c.Liked = true
			}
		}
	}
	return comments, total, nil
}

func (s *videoServiceImpl) ToggleCommentLike(ctx context.Context, userUUID, commentUUID string) (bool, int64, error) {
	comment, err := s.commentRepo.FindRootByUUID(ctx, commentUUID)
	targetIsRoot := true
	if err != nil {
		return false, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	if comment == nil {
		if c2, err2 := s.commentRepo.FindReplyByUUID(ctx, commentUUID); err2 != nil {
			return false, 0, errno.NewBizError(errno.ErrDatabase, err2)
		} else if c2 != nil {
			comment = c2
			targetIsRoot = false
		}
	}
	if comment == nil {
		return false, 0, errno.ErrNotFound
	}
	liked := false
	if s.commentLikeRepo != nil {
		if ok, err := s.commentLikeRepo.Exists(ctx, commentUUID, userUUID); err == nil && ok {
			liked = true
		}
	}
	if liked {
		if err := s.commentLikeRepo.Remove(ctx, commentUUID, userUUID); err != nil {
			return false, 0, errno.NewBizError(errno.ErrDatabase, err)
		}
	} else {
		if _, err := s.commentLikeRepo.Add(ctx, commentUUID, userUUID); err != nil {
			return false, 0, errno.NewBizError(errno.ErrDatabase, err)
		}
	}
	likeCount, err := s.commentLikeRepo.CountByComment(ctx, commentUUID)
	if err != nil {
		return !liked, 0, errno.NewBizError(errno.ErrDatabase, err)
	}
	if targetIsRoot {
		if err := s.commentRepo.UpdateRootLikeCount(ctx, commentUUID, likeCount); err != nil {
			logger.Warnf("sync root comment like_count failed comment=%s err=%v", commentUUID, err)
		}
	} else {
		if err := s.commentRepo.UpdateReplyLikeCount(ctx, commentUUID, likeCount); err != nil {
			logger.Warnf("sync reply like_count failed comment=%s err=%v", commentUUID, err)
		}
	}
	return !liked, likeCount, nil
}
