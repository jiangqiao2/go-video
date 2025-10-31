package resource

import (
	"sync"
	"transcode-service/pkg/assert"
	"transcode-service/pkg/config"
	"transcode-service/pkg/manager"
	"transcode-service/pkg/repository"

	"gorm.io/gorm"
)

var (
	mysqlResourceOnce      sync.Once
	singletonMysqlResource *MySqlResource
)

// MySqlResource MySQL资源管理器
type MySqlResource struct {
	db *repository.Database
}

// DefaultMysqlResource 获取MySQL资源单例
func DefaultMysqlResource() *MySqlResource {
	assert.NotCircular()
	mysqlResourceOnce.Do(func() {
		singletonMysqlResource = &MySqlResource{}
	})
	assert.NotNil(singletonMysqlResource)
	return singletonMysqlResource
}

// MustOpen 打开MySQL连接
func (r *MySqlResource) MustOpen() {
	if r.db == nil {
		// 从全局配置获取数据库配置
		cfg := config.GetGlobalConfig()
		if cfg == nil {
			panic("global config not initialized")
		}

		db, err := repository.NewDatabase(&cfg.Database)
		if err != nil {
			panic("failed to create database: " + err.Error())
		}
		// if err := db.Self.AutoMigrate(&po.TranscodeTask{}); err != nil {
		// 	panic("failed to migrate database schema: " + err.Error())
		// }
		r.db = db
	}
	assert.NotNil(r.db)
}

// Close 关闭MySQL连接
func (r *MySqlResource) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

// MainDB 获取主数据库连接
func (r *MySqlResource) MainDB() *gorm.DB {
	return r.db.Self
}

// ReadDB 获取读数据库连接（当前与主库相同）
func (r *MySqlResource) ReadDB() *gorm.DB {
	return r.db.Self
}

// MySqlResourcePlugin MySQL资源插件
type MySqlResourcePlugin struct{}

// Name 返回插件名称
func (p *MySqlResourcePlugin) Name() string {
	return "mysql"
}

// MustCreateResource 创建MySQL资源
func (p *MySqlResourcePlugin) MustCreateResource() manager.Resource {
	return DefaultMysqlResource()
}

// NewMysqlResource 创建MySQL资源实例（支持依赖注入）
func NewMysqlResource(db *gorm.DB) *gorm.DB {
	return db
}
