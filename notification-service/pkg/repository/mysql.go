package repository

import (
	"fmt"
	"log"
	"os"
	"time"

	"notification-service/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps primary gorm DB instance.
type Database struct {
	Self *gorm.DB
}

// NewDatabase initialises a Database from config.
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	selfDB, err := initSelfDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return &Database{Self: selfDB}, nil
}

// Close closes underlying sql.DB if possible.
func (db *Database) Close() {
	if db == nil || db.Self == nil {
		return
	}
	if sqlDB, err := db.Self.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func initSelfDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	loggerWriter := log.New(os.Stdout, "\r\n", log.LstdFlags)
	gormLogger := logger.New(
		loggerWriter,
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

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

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
	return gormDB, nil
}
