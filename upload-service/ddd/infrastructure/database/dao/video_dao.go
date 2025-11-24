package dao

import (
    "context"
    "errors"
    "strings"

    log "github.com/sirupsen/logrus"
    "gorm.io/gorm"
    "github.com/google/uuid"

	"upload-service/ddd/infrastructure/database/po"
	"upload-service/internal/resource"
)

// VideoDao encapsulates CRUD operations for video_publish table.
type VideoDao struct {
	db *gorm.DB
}

// NewVideoDao creates a dao backed by the default mysql resource.
func NewVideoDao() *VideoDao {
	return &VideoDao{
		db: resource.DefaultMysqlResource().MainDB(),
	}
}

func (d *VideoDao) Create(ctx context.Context, video *po.VideoPo) error {
    return d.db.WithContext(ctx).Model(&po.VideoPo{}).Create(video).Error
}

func (d *VideoDao) CreateWithTags(ctx context.Context, video *po.VideoPo, tags []string) error {
    return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Model(&po.VideoPo{}).Create(video).Error; err != nil {
            return err
        }
        if len(tags) == 0 {
            return nil
        }
        normalized := make([]string, 0, len(tags))
        for _, t := range tags {
            tt := strings.TrimSpace(t)
            if tt != "" {
                normalized = append(normalized, tt)
            }
        }
        for _, name := range normalized {
            code := toTagCode(name)
            var tag po.Tag
            err := tx.Model(&po.Tag{}).Where("code = ? AND is_deleted = 0", code).First(&tag).Error
            if err != nil {
                if errors.Is(err, gorm.ErrRecordNotFound) {
                    tag = po.Tag{Name: name, Code: code, TagUUID: uuid.NewString()}
                    if err := tx.Model(&po.Tag{}).Create(&tag).Error; err != nil {
                        return err
                    }
                } else {
                    return err
                }
            }
            var cnt int64
            if err := tx.Model(&po.VideoTag{}).Where("video_uuid = ? AND tag_uuid = ? AND is_deleted = 0", video.VideoUUID, tag.TagUUID).Count(&cnt).Error; err != nil {
                return err
            }
            if cnt > 0 {
                continue
            }
            vt := &po.VideoTag{VideoUUID: video.VideoUUID, TagUUID: tag.TagUUID}
            if err := tx.Model(&po.VideoTag{}).Create(vt).Error; err != nil {
                return err
            }
        }
        return nil
    })
}

func toTagCode(s string) string {
    s = strings.TrimSpace(s)
    r := make([]rune, 0, len(s))
    for _, ch := range s {
        switch ch {
        case ' ', '\t', '\n', '\r', '-':
            r = append(r, '_')
        default:
            r = append(r, ch)
        }
    }
    return strings.ToLower(string(r))
}

func (d *VideoDao) QueryByUploadVideoUUID(ctx context.Context, uploadVideoUUID string) (*po.VideoPo, error) {
	var video po.VideoPo
	err := d.db.WithContext(ctx).
		Model(&po.VideoPo{}).
		Where("upload_video_uuid = ? AND is_deleted = 0", uploadVideoUUID).
		First(&video).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("QueryByUploadVideoUUID failed: %v", err)
		return nil, err
	}
	return &video, nil
}

func (d *VideoDao) QueryByVideoUUID(ctx context.Context, videoUUID string) (*po.VideoPo, error) {
	var video po.VideoPo
	err := d.db.WithContext(ctx).
		Model(&po.VideoPo{}).
		Where("video_uuid = ? AND is_deleted = 0", videoUUID).
		First(&video).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("QueryByVideoUUID failed: %v", err)
		return nil, err
	}
	return &video, nil
}

func (d *VideoDao) UpdateTranscodeInfo(ctx context.Context, videoUUID string, values map[string]interface{}) error {
	return d.db.WithContext(ctx).
		Model(&po.VideoPo{}).
		Where("video_uuid = ? AND is_deleted = 0", videoUUID).
		Updates(values).Error
}

func (d *VideoDao) QueryByUser(ctx context.Context, userUUID string, status string, offset, limit int) ([]*po.VideoPo, int64, error) {
    var videos []*po.VideoPo
    query := d.db.WithContext(ctx).
        Model(&po.VideoPo{}).
        Where("user_uuid = ? AND is_deleted = 0", userUUID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Errorf("QueryByUser count failed: %v", err)
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Order("created_at DESC").Find(&videos).Error; err != nil {
		log.Errorf("QueryByUser find failed: %v", err)
		return nil, 0, err
	}

    return videos, total, nil
}

func (d *VideoDao) QueryByStatus(ctx context.Context, status string, offset, limit int) ([]*po.VideoPo, int64, error) {
    var videos []*po.VideoPo
    query := d.db.WithContext(ctx).
        Model(&po.VideoPo{}).
        Where("is_deleted = 0")

    if status != "" {
        query = query.Where("status = ?", status)
    }

    var total int64
    if err := query.Count(&total).Error; err != nil {
        log.Errorf("QueryByStatus count failed: %v", err)
        return nil, 0, err
    }

    if limit > 0 {
        query = query.Offset(offset).Limit(limit)
    }

    if err := query.Order("published_at DESC").Find(&videos).Error; err != nil {
        log.Errorf("QueryByStatus find failed: %v", err)
        return nil, 0, err
    }

    return videos, total, nil
}
