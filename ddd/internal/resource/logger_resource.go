package resource

import (
	"sync"

	"go-video/pkg/assert"
	"go-video/pkg/config"
	"go-video/pkg/logger"
	"go-video/pkg/manager"
)

var (
	loggerResourceOnce      sync.Once
	singletonLoggerResource *LoggerResource
)

// LoggerResource 日志资源管理器
type LoggerResource struct {
	logger *logger.Logger
}

// DefaultLoggerResource 获取日志资源单例
func DefaultLoggerResource() *LoggerResource {
	assert.NotCircular()
	loggerResourceOnce.Do(func() {
		singletonLoggerResource = &LoggerResource{}
	})
	assert.NotNil(singletonLoggerResource)
	return singletonLoggerResource
}

// MustOpen 初始化日志服务
func (r *LoggerResource) MustOpen() {
	// 日志服务已经在app.go中初始化，这里只是标记资源已打开
	// 不需要重复初始化，避免冲突
	if r.logger == nil {
		// 创建一个默认logger实例用于资源管理
		r.logger = logger.DefaultLogger()
	}
	assert.NotNil(r.logger)
}

// Close 关闭日志服务
func (r *LoggerResource) Close() {
	if r.logger != nil {
		r.logger.Close()
	}
}

// GetLogger 获取日志器
func (r *LoggerResource) GetLogger() *logger.Logger {
	return r.logger
}

// LoggerResourcePlugin 日志资源插件
type LoggerResourcePlugin struct{}

// Name 返回插件名称
func (p *LoggerResourcePlugin) Name() string {
	return "logger"
}

// MustCreateResource 创建日志资源
func (p *LoggerResourcePlugin) MustCreateResource() manager.Resource {
	return DefaultLoggerResource()
}

// NewLoggerResource 创建日志资源实例（支持依赖注入）
func NewLoggerResource(cfg *config.Config) *logger.Logger {
	return logger.NewLogger(cfg)
}
