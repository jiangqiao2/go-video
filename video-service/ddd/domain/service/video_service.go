package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"video-service/ddd/application/cqe"
	"video-service/ddd/application/dto"
)

// VideoService is a lightweight in-memory implementation to bootstrap the service.
type VideoService interface {
	Publish(req *cqe.PublishVideoReq) (*dto.VideoDto, error)
	Get(videoUUID string) (*dto.VideoDto, error)
	List(page, size int) ([]*dto.VideoDto, int64)
	Like(userUUID, videoUUID string) error
	Unlike(userUUID, videoUUID string) error
	Play(videoUUID string) error
	AddComment(req *cqe.CommentCreateReq) (*dto.CommentDto, error)
	ListComments(videoUUID string, page, size int) ([]*dto.CommentDto, int64)
}

func NewInMemoryVideoService() VideoService {
	return &memoryVideoService{
		videos:   make(map[string]*dto.VideoDto),
		likes:    make(map[string]map[string]bool),
		comments: make(map[string][]*dto.CommentDto),
	}
}

type memoryVideoService struct {
	mu       sync.Mutex
	videos   map[string]*dto.VideoDto
	likes    map[string]map[string]bool // videoUUID -> set of userUUID
	comments map[string][]*dto.CommentDto
}

func (m *memoryVideoService) Publish(req *cqe.PublishVideoReq) (*dto.VideoDto, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	video := &dto.VideoDto{
		VideoUUID:   req.VideoUUID,
		UploadVideo: req.UploadVideoUUID,
		UserUUID:    req.UserUUID,
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		VideoURL:    req.VideoURL,
		Status:      "Published",
		Tags:        req.Tags,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	m.videos[video.VideoUUID] = video
	return video, nil
}

func (m *memoryVideoService) Get(videoUUID string) (*dto.VideoDto, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.videos[videoUUID]
	if !ok {
		return nil, errors.New("video not found")
	}
	return v, nil
}

func (m *memoryVideoService) List(page, size int) ([]*dto.VideoDto, int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	start := (page - 1) * size
	idx := 0
	total := int64(len(m.videos))
	res := make([]*dto.VideoDto, 0, size)
	for _, v := range m.videos {
		if idx >= start && len(res) < size {
			res = append(res, v)
		}
		idx++
	}
	return res, total
}

func (m *memoryVideoService) Like(userUUID, videoUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.videos[videoUUID]; !ok {
		return errors.New("video not found")
	}
	if m.likes[videoUUID] == nil {
		m.likes[videoUUID] = make(map[string]bool)
	}
	if !m.likes[videoUUID][userUUID] {
		m.likes[videoUUID][userUUID] = true
		m.videos[videoUUID].LikeCount++
	}
	return nil
}

func (m *memoryVideoService) Unlike(userUUID, videoUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.videos[videoUUID]; !ok {
		return errors.New("video not found")
	}
	if m.likes[videoUUID] != nil && m.likes[videoUUID][userUUID] {
		delete(m.likes[videoUUID], userUUID)
		if m.videos[videoUUID].LikeCount > 0 {
			m.videos[videoUUID].LikeCount--
		}
	}
	return nil
}

func (m *memoryVideoService) Play(videoUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.videos[videoUUID]; ok {
		v.PlayCount++
		return nil
	}
	return errors.New("video not found")
}

func (m *memoryVideoService) AddComment(req *cqe.CommentCreateReq) (*dto.CommentDto, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.videos[req.VideoUUID]; !ok {
		return nil, errors.New("video not found")
	}
	c := &dto.CommentDto{
		CommentUUID: fmt.Sprintf("%s-%d", req.VideoUUID, len(m.comments[req.VideoUUID])+1),
		VideoUUID:   req.VideoUUID,
		UserUUID:    req.UserUUID,
		Content:     req.Content,
		CreatedAt:   time.Now().Format(time.RFC3339),
		ParentUUID:  req.ParentUUID,
	}
	m.comments[req.VideoUUID] = append(m.comments[req.VideoUUID], c)
	return c, nil
}

func (m *memoryVideoService) ListComments(videoUUID string, page, size int) ([]*dto.CommentDto, int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	comments := m.comments[videoUUID]
	total := int64(len(comments))
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	start := (page - 1) * size
	end := start + size
	if start > len(comments) {
		return []*dto.CommentDto{}, total
	}
	if end > len(comments) {
		end = len(comments)
	}
	return comments[start:end], total
}
