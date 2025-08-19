package database

import (
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"go-video/pkg/config"
)

var (
	// 数据库连接单例
	dbOnce      sync.Once
	singletonDB *gorm.DB
)

// DefaultDB 返回默认数据库连接单例
func DefaultDB() *gorm.DB {
	dbOnce.Do(func() {
		cfg, err := config.Load("configs/config.dev.yaml")
		if err != nil {
			panic("failed to load config: " + err.Error())
		}

		db, err := gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{})
		if err != nil {
			panic("failed to connect database: " + err.Error())
		}
		singletonDB = db
	})
	if singletonDB == nil {
		panic("failed to create database singleton")
	}
	return singletonDB
}

// GetDB 获取数据库连接（别名，为了兼容性）
func GetDB() *gorm.DB {
	return DefaultDB()
}

// InitDB 初始化数据库连接（支持依赖注入）
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db, nil
}