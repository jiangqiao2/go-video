package persistence

import (
    "context"
    "upload-service/ddd/domain/repo"
    "upload-service/ddd/infrastructure/database/dao"
    "upload-service/ddd/infrastructure/database/po"
)

type tagRepositoryImpl struct {
    dao *dao.TagDao
}

func NewTagRepository() repo.TagRepository {
    return &tagRepositoryImpl{dao: dao.NewTagDao()}
}

func (r *tagRepositoryImpl) ListAll(ctx context.Context) ([]*po.Tag, error) {
    return r.dao.ListAll(ctx)
}