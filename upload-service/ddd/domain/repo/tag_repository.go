package repo

import (
	"context"
	"upload-service/ddd/infrastructure/database/po"
)

type TagRepository interface {
	ListAll(ctx context.Context) ([]*po.Tag, error)
}
