package persistence

import "transcode-service/ddd/infrastructure/database/dao"

type transcodeRepositoryImpl struct {
	transTaskDao *dao.TranscodeTaskDAO
}
