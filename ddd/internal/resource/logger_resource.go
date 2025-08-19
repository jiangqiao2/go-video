package resource

import (
	"sync"

	"go-video/pkg/assert"
	"go-video/pkg/config"
	"go-video/pkg/logger"
)

var (
	loggerResourceOnce      sync.Once
	singletonLoggerResource *logger.Logger
)

// DefaultLoggerResource 获取日志资源单例
func DefaultLoggerResource() *logger.Logger {
	assert.NotCircular()
	loggerResourceOnce.Do(func() {
		singletonLoggerResource = logger.DefaultLogger()
	})
	assert.NotNil(singletonLoggerResource)
	return singletonLoggerResource
}

// NewLoggerResource 创建日志资源实例（支持依赖注入）
func NewLoggerResource(cfg *config.Config) *logger.Logger {
	return logger.NewLogger(cfg)
}