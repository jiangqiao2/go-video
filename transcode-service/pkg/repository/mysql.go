package repository

import (
	"fmt"
	"log"
	"os"
	"time"
	"transcode-service/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 便于多个数据源扩展
type Database struct {
	Self *gorm.DB
}

// NewDatabase 初始化数据库连接
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	selfDB, err := initSelfDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Database{
		Self: selfDB,
	}, nil
}

// NewDatabaseTx 用给定的tx创建Database对象
func NewDatabaseTx(tx *gorm.DB) *Database {
	return &Database{
		Self: tx,
	}
}

// Close 关闭数据库连接
func (db *Database) Close() {
	if sqlDB, err := db.Self.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

// initSelfDB 初始化数据库连接
func initSelfDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// 使用配置中的GetDSN方法构建连接字符串
	dsn := cfg.GetDSN()

	// 配置日志输出
	loggerWriter := log.New(os.Stdout, "\r\n", log.LstdFlags)

	// 配置GORM日志
	gormLogger := logger.New(
		loggerWriter,
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 打开数据库连接
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{
		CreateBatchSize:        1000,
		SkipDefaultTransaction: false,
		Logger:                 gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100) // 默认值
	}

	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10) // 默认值
	}

	// 设置连接最大生存时间
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(time.Hour) // 默认值
	}

	return gormDB, nil
}
