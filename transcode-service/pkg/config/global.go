package config

import (
	"sync"
)

var (
	// 全局配置实例
	globalConfig      *Config
	globalConfigMutex sync.RWMutex
)

// SetGlobalConfig 设置全局配置
func SetGlobalConfig(cfg *Config) {
	globalConfigMutex.Lock()
	defer globalConfigMutex.Unlock()
	globalConfig = cfg
}

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *Config {
	globalConfigMutex.RLock()
	defer globalConfigMutex.RUnlock()
	return globalConfig
}

// IsGlobalConfigInitialized 检查全局配置是否已初始化
func IsGlobalConfigInitialized() bool {
	globalConfigMutex.RLock()
	defer globalConfigMutex.RUnlock()
	return globalConfig != nil
}
