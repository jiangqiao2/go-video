package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server          ServerConfig          `mapstructure:"server"`
	Database        DatabaseConfig        `mapstructure:"database"`
	Redis           RedisConfig           `mapstructure:"redis"`
	Kafka           KafkaConfig           `mapstructure:"kafka"`
	JWT             JWTConfig             `mapstructure:"jwt"`
	Log             LogConfig             `mapstructure:"log"`
	Minio           MinioConfig           `mapstructure:"minio"`
	RustFS          RustFSConfig          `mapstructure:"rustfs"`
	Transcode       TranscodeConfig       `mapstructure:"transcode"`
	Worker          WorkerConfig          `mapstructure:"worker"`
	Scheduler       SchedulerConfig       `mapstructure:"scheduler"`
	ServiceRegistry ServiceRegistryConfig `mapstructure:"service_registry"`
	GRPCServer      GRPCServerConfig      `mapstructure:"grpc_server"`
	GRPCClient      GRPCClientConfig      `mapstructure:"grpc_client"`
	Dependencies    DependenciesConfig    `mapstructure:"dependencies"`
	Public          PublicConfig          `mapstructure:"public"`
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
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	EnableTLS    bool          `mapstructure:"enable_tls"`
}

// ServiceRegistryConfig registration configuration.
type ServiceRegistryConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	ServiceName     string        `mapstructure:"service_name"`
	ServiceID       string        `mapstructure:"service_id"`
	RegisterHost    string        `mapstructure:"register_host"`
	TTL             time.Duration `mapstructure:"ttl"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
}

// GRPCServerConfig gRPC server configuration.
type GRPCServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// GRPCClientConfig defines outbound gRPC client behaviour.
type GRPCClientConfig struct {
	Timeout        time.Duration `mapstructure:"timeout"`
	MaxRecvMsgSize int           `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int           `mapstructure:"max_send_msg_size"`
	RetryTimes     int           `mapstructure:"retry_times"`
}

// DependenciesConfig enumerates downstream services used by transcode-service.
type DependenciesConfig struct {
	UploadService UploadServiceConfig `mapstructure:"upload_service"`
	VideoService  VideoServiceConfig  `mapstructure:"video_service"`
}

// UploadServiceConfig describes upload-service discovery metadata.
type UploadServiceConfig struct {
	ServiceName string        `mapstructure:"service_name"`
	Address     string        `mapstructure:"address"`
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

type VideoServiceConfig struct {
	ServiceName string        `mapstructure:"service_name"`
	Address     string        `mapstructure:"address"`
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Timeout     time.Duration `mapstructure:"timeout"`
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

// RustFSConfig RustFS配置
type RustFSConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// PublicConfig 对外访问配置
type PublicConfig struct {
	StorageBase string `mapstructure:"storage_base"`
}

// TranscodeConfig 转码配置
type TranscodeConfig struct {
	FFmpeg         FFmpegConfig   `mapstructure:"ffmpeg"`
	OutputFormats  []OutputFormat `mapstructure:"output_formats"`
	SkipFullUpload bool           `mapstructure:"skip_full_upload"`
}

// OutputFormat 输出格式配置
type OutputFormat struct {
	Name       string `mapstructure:"name"`
	Resolution string `mapstructure:"resolution"`
	Bitrate    string `mapstructure:"bitrate"`
	Codec      string `mapstructure:"codec"`
	Preset     string `mapstructure:"preset"`
}

// FFmpegConfig FFmpeg相关配置
type FFmpegConfig struct {
	BinaryPath         string        `mapstructure:"binary_path"`
	TempDir            string        `mapstructure:"temp_dir"`
	MaxConcurrentTasks int           `mapstructure:"max_concurrent_tasks"`
	Timeout            time.Duration `mapstructure:"timeout"`
	VideoCodec         string        `mapstructure:"video_codec"`
	HardwareAccel      string        `mapstructure:"hardware_accel"`
	VideoPreset        string        `mapstructure:"video_preset"`
	Threads            int           `mapstructure:"threads"`
	UseHardwareDecode  bool          `mapstructure:"use_hardware_decode"`
	DecoderThreads     int           `mapstructure:"decoder_threads"`
	CuvidSurfaces      int           `mapstructure:"cuvid_surfaces"`
}

// WorkerConfig Worker相关配置
type WorkerConfig struct {
	Enabled               bool          `mapstructure:"enabled"`
	WorkerID              string        `mapstructure:"worker_id"`
	HeartbeatInterval     time.Duration `mapstructure:"heartbeat_interval"`
	TaskPollInterval      time.Duration `mapstructure:"task_poll_interval"`
	MaxConcurrentTasks    int           `mapstructure:"max_concurrent_tasks"`
	HLSMaxConcurrentTasks int           `mapstructure:"hls_max_concurrent_tasks"`
	QueueCapacity         int           `mapstructure:"queue_capacity"`
	ShutdownGracePeriod   time.Duration `mapstructure:"shutdown_grace_period"`
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
	Secret                string        `mapstructure:"secret"`
	Issuer                string        `mapstructure:"issuer"`
	RSAPrivateKeyPath     string        `mapstructure:"rsa_private_key_path"`
	RSAPublicKeyPath      string        `mapstructure:"rsa_public_key_path"`
	RSAPrivateKeyPassword string        `mapstructure:"rsa_private_key_password"`
	ExpireTime            time.Duration `mapstructure:"expire_time"`
	RefreshExpireTime     time.Duration `mapstructure:"refresh_expire_time"`
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

	// 保持向后兼容：默认开启服务注册，可配置关闭
	viper.SetDefault("service_registry.enabled", true)
	viper.SetDefault("dependencies.upload_service.service_name", "upload-service")
	viper.SetDefault("kafka.enabled", true)
	viper.SetDefault("kafka.client_id", "transcode-service")
	viper.SetDefault("kafka.group_id", "transcode-service-group")
	viper.SetDefault("kafka.bootstrap_servers", []string{"localhost:29092"})
	viper.SetDefault("kafka.topics.transcode_tasks", "transcode.tasks")
	viper.SetDefault("kafka.commit_on_decode_error", true)
	viper.SetDefault("kafka.commit_on_process_error", false)

	// 设置环境变量前缀
	viper.SetEnvPrefix("GO_VIDEO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
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

	// RustFS默认端口
	if c.RustFS.Endpoint == "" {
		c.RustFS.Endpoint = c.Minio.Endpoint
	}

	// Worker相关默认值
	if c.Worker.MaxConcurrentTasks <= 0 {
		if c.Transcode.FFmpeg.MaxConcurrentTasks > 0 {
			c.Worker.MaxConcurrentTasks = c.Transcode.FFmpeg.MaxConcurrentTasks
		} else {
			c.Worker.MaxConcurrentTasks = 2
		}
	}
	if c.Worker.HLSMaxConcurrentTasks <= 0 {
		c.Worker.HLSMaxConcurrentTasks = c.Worker.MaxConcurrentTasks
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
	if c.Transcode.FFmpeg.VideoCodec == "" {
		c.Transcode.FFmpeg.VideoCodec = "libx264"
	}
	if c.Transcode.FFmpeg.VideoPreset == "" {
		c.Transcode.FFmpeg.VideoPreset = "medium"
	}
	if c.Transcode.FFmpeg.Threads < 0 {
		c.Transcode.FFmpeg.Threads = 0
	}
	if c.Transcode.FFmpeg.Timeout == 0 {
		c.Transcode.FFmpeg.Timeout = time.Hour
	}
	if c.GRPCServer.Host == "" {
		c.GRPCServer.Host = "0.0.0.0"
	}
	if c.GRPCServer.Port == 0 {
		c.GRPCServer.Port = 9092
	}
	if c.ServiceRegistry.TTL == 0 {
		c.ServiceRegistry.TTL = 30 * time.Second
	}
	if c.ServiceRegistry.RefreshInterval == 0 {
		c.ServiceRegistry.RefreshInterval = 10 * time.Second
	}
	if c.GRPCClient.Timeout <= 0 {
		c.GRPCClient.Timeout = 30 * time.Second
	}
	if c.GRPCClient.RetryTimes < 0 {
		c.GRPCClient.RetryTimes = 0
	}
	if c.Dependencies.UploadService.ServiceName == "" {
		c.Dependencies.UploadService.ServiceName = "upload-service"
	}
	if c.Dependencies.UploadService.Port <= 0 {
		c.Dependencies.UploadService.Port = 9093
	}
	if c.Dependencies.UploadService.Timeout <= 0 {
		c.Dependencies.UploadService.Timeout = c.GRPCClient.Timeout
	}
	if c.Dependencies.VideoService.ServiceName == "" {
		c.Dependencies.VideoService.ServiceName = "video-service"
	}
	if c.Dependencies.VideoService.Port <= 0 {
		c.Dependencies.VideoService.Port = 9094
	}
	if c.Dependencies.VideoService.Timeout <= 0 {
		c.Dependencies.VideoService.Timeout = c.GRPCClient.Timeout
	}
	if len(c.Kafka.BootstrapServers) == 0 {
		c.Kafka.BootstrapServers = []string{"localhost:29092"}
	}
	if c.Kafka.ClientID == "" {
		c.Kafka.ClientID = "transcode-service"
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

// KafkaConfig Kafka配置
type KafkaConfig struct {
	BootstrapServers     []string          `mapstructure:"bootstrap_servers"`
	ClientID             string            `mapstructure:"client_id"`
	GroupID              string            `mapstructure:"group_id"`
	Enabled              bool              `mapstructure:"enabled"`
	Topics               KafkaTopicsConfig `mapstructure:"topics"`
	CommitOnDecodeError  bool              `mapstructure:"commit_on_decode_error"`
	CommitOnProcessError bool              `mapstructure:"commit_on_process_error"`
}

type KafkaTopicsConfig struct {
	TranscodeTasks string `mapstructure:"transcode_tasks"`
}
