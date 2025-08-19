package resource

import (
	"sync"

	"gorm.io/gorm"
	"go-video/pkg/assert"
	"go-video/pkg/database"
)

var (
	mysqlResourceOnce      sync.Once
	singletonMysqlResource *gorm.DB
)

// DefaultMysqlResource 获取MySQL资源单例
func DefaultMysqlResource() *gorm.DB {
	assert.NotCircular()
	mysqlResourceOnce.Do(func() {
		singletonMysqlResource = database.DefaultDB()
	})
	assert.NotNil(singletonMysqlResource)
	return singletonMysqlResource
}

// NewMysqlResource 创建MySQL资源实例（支持依赖注入）
func NewMysqlResource(db *gorm.DB) *gorm.DB {
	return db
}