package resource

import (
	"sync"

	"gorm.io/gorm"
)

var (
	mainDB *gorm.DB
	once   sync.Once
)

// SetMainDB sets the global main DB instance for this service.
// It should be called once during startup in app.Run.
func SetMainDB(db *gorm.DB) {
	if db == nil {
		panic("SetMainDB called with nil db")
	}
	once.Do(func() {
		mainDB = db
	})
}

// MainDB returns the main DB instance. It panics if not initialised.
func MainDB() *gorm.DB {
	if mainDB == nil {
		panic("MainDB not initialized; call resource.SetMainDB in app.Run first")
	}
	return mainDB
}
