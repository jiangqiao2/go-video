package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	Minio     MinioConfig     `mapstructure:"minio"`
	Transcode TranscodeConfig `mapstructure:"transcode"`
	Worker    WorkerConfig    `mapstructure:"worker"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// MinioConfig MinIO配置
type MinioConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKey       string `mapstructure:"access_key"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	SecretKey       string `mapstructure:"secret_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

// TranscodeConfig 转码配置
type TranscodeConfig struct {
	FFmpeg FFmpegConfig `mapstructure:"ffmpeg"`
}

// FFmpegConfig FFmpeg相关配置
type FFmpegConfig struct {
	BinaryPath         string        `mapstructure:"binary_path"`
	TempDir            string        `mapstructure:"temp_dir"`
	MaxConcurrentTasks int           `mapstructure:"max_concurrent_tasks"`
	Timeout            time.Duration `mapstructure:"timeout"`
}

// WorkerConfig Worker相关配置
type WorkerConfig struct {
	Enabled             bool          `mapstructure:"enabled"`
	WorkerID            string        `mapstructure:"worker_id"`
	HeartbeatInterval   time.Duration `mapstructure:"heartbeat_interval"`
	TaskPollInterval    time.Duration `mapstructure:"task_poll_interval"`
	MaxConcurrentTasks  int           `mapstructure:"max_concurrent_tasks"`
	QueueCapacity       int           `mapstructure:"queue_capacity"`
	ShutdownGracePeriod time.Duration `mapstructure:"shutdown_grace_period"`
}

// SchedulerConfig 调度器相关配置
type SchedulerConfig struct {
	Enabled                bool          `mapstructure:"enabled"`
	TaskPollInterval       time.Duration `mapstructure:"task_poll_interval"`
	WorkerHeartbeatTimeout time.Duration `mapstructure:"worker_heartbeat_timeout"`
	MaxRetryCount          int           `mapstructure:"max_retry_count"`
	CleanupInterval        time.Duration `mapstructure:"cleanup_interval"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	ExpireTime        time.Duration `mapstructure:"expire_time"`
	RefreshExpireTime time.Duration `mapstructure:"refresh_expire_time"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置环境变量前缀
	viper.SetEnvPrefix("GO_VIDEO")
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 解析配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	config.normalize()

	return &config, nil
}

// normalize 补全配置的默认值
func (c *Config) normalize() {
	// 兼容不同的密钥字段
	if c.Minio.AccessKeyID == "" {
		c.Minio.AccessKeyID = c.Minio.AccessKey
	}
	if c.Minio.SecretAccessKey == "" {
		c.Minio.SecretAccessKey = c.Minio.SecretKey
	}

	// Worker相关默认值
	if c.Worker.MaxConcurrentTasks <= 0 {
		if c.Transcode.FFmpeg.MaxConcurrentTasks > 0 {
			c.Worker.MaxConcurrentTasks = c.Transcode.FFmpeg.MaxConcurrentTasks
		} else {
			c.Worker.MaxConcurrentTasks = 2
		}
	}
	if c.Worker.QueueCapacity <= 0 {
		c.Worker.QueueCapacity = c.Worker.MaxConcurrentTasks * 10
		if c.Worker.QueueCapacity <= 0 {
			c.Worker.QueueCapacity = 100
		}
	}
	if c.Worker.ShutdownGracePeriod == 0 {
		c.Worker.ShutdownGracePeriod = 10 * time.Second
	}

	// FFmpeg临时目录默认值
	if c.Transcode.FFmpeg.TempDir == "" {
		c.Transcode.FFmpeg.TempDir = "/tmp/transcode"
	}
	if c.Transcode.FFmpeg.BinaryPath == "" {
		c.Transcode.FFmpeg.BinaryPath = "ffmpeg"
	}
	if c.Transcode.FFmpeg.Timeout == 0 {
		c.Transcode.FFmpeg.Timeout = time.Hour
	}
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetMinioEndpoint 获取MinIO端点
func (c *MinioConfig) GetMinioEndpoint() string {
	return c.Endpoint
}
