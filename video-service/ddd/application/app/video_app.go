package app

import (
	"video-service/ddd/application/cqe"
	"video-service/ddd/application/dto"
	"video-service/ddd/domain/service"

	"github.com/google/uuid"
)

type VideoApp interface {
	Publish(req *cqe.PublishVideoReq) (*dto.VideoDto, error)
	Get(videoUUID string) (*dto.VideoDto, error)
	List(page, size int) (*dto.VideoListDto, error)
	Like(userUUID, videoUUID string) error
	Unlike(userUUID, videoUUID string) error
	Play(videoUUID string) error
	AddComment(req *cqe.CommentCreateReq) (*dto.CommentDto, error)
	ListComments(videoUUID string, page, size int) (*dto.CommentListDto, error)
}

type videoAppImpl struct {
	svc service.VideoService
}

func NewVideoApp() VideoApp {
	return &videoAppImpl{
		svc: service.NewInMemoryVideoService(),
	}
}

func (v *videoAppImpl) Publish(req *cqe.PublishVideoReq) (*dto.VideoDto, error) {
	if req.VideoUUID == "" {
		req.VideoUUID = uuid.NewString()
	}
	return v.svc.Publish(req)
}

func (v *videoAppImpl) Get(videoUUID string) (*dto.VideoDto, error) {
	return v.svc.Get(videoUUID)
}

func (v *videoAppImpl) List(page, size int) (*dto.VideoListDto, error) {
	list, total := v.svc.List(page, size)
	return &dto.VideoListDto{
		List:  list,
		Page:  page,
		Size:  size,
		Total: total,
	}, nil
}

func (v *videoAppImpl) Like(userUUID, videoUUID string) error {
	return v.svc.Like(userUUID, videoUUID)
}

func (v *videoAppImpl) Unlike(userUUID, videoUUID string) error {
	return v.svc.Unlike(userUUID, videoUUID)
}

func (v *videoAppImpl) Play(videoUUID string) error {
	return v.svc.Play(videoUUID)
}

func (v *videoAppImpl) AddComment(req *cqe.CommentCreateReq) (*dto.CommentDto, error) {
	return v.svc.AddComment(req)
}

func (v *videoAppImpl) ListComments(videoUUID string, page, size int) (*dto.CommentListDto, error) {
	list, total := v.svc.ListComments(videoUUID, page, size)
	return &dto.CommentListDto{
		List:  list,
		Page:  page,
		Size:  size,
		Total: total,
	}, nil
}
