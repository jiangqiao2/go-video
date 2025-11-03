package config

import "sync"

var (
	globalConfig      *Config
	globalConfigMutex sync.RWMutex
)

// SetGlobalConfig stores the configuration for later retrieval by shared packages.
func SetGlobalConfig(cfg *Config) {
	globalConfigMutex.Lock()
	defer globalConfigMutex.Unlock()
	globalConfig = cfg
}

// GetGlobalConfig returns the previously stored configuration instance.
func GetGlobalConfig() *Config {
	globalConfigMutex.RLock()
	defer globalConfigMutex.RUnlock()
	return globalConfig
}

// IsGlobalConfigInitialized reports whether SetGlobalConfig has been called.
func IsGlobalConfigInitialized() bool {
	globalConfigMutex.RLock()
	defer globalConfigMutex.RUnlock()
	return globalConfig != nil
}

