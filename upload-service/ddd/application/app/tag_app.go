package app

import (
	"context"
	"upload-service/ddd/application/cqe"
	"upload-service/ddd/application/dto"
	"upload-service/ddd/infrastructure/database/persistence"
	"upload-service/ddd/infrastructure/database/po"
)

type TagApp interface {
	ListTags(ctx context.Context, req *cqe.ListTagsReq) (*dto.TagListDto, error)
}

type tagAppImpl struct{}

func DefaultTagApp() TagApp { return &tagAppImpl{} }

func (a *tagAppImpl) ListTags(ctx context.Context, req *cqe.ListTagsReq) (*dto.TagListDto, error) {
	repo := persistence.NewTagRepository()
	var list []*po.Tag
	var err error
	list, err = repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]*dto.TagDto, 0, len(list))
	for _, t := range list {
		dtos = append(dtos, &dto.TagDto{Id: t.Id, TagUUID: t.TagUUID, Name: t.Name, Code: t.Code, Description: t.Description})
	}
	return dto.NewTagListDto(dtos), nil
}
