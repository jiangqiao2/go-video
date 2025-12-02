package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server              ServerConfig          `mapstructure:"server"`
	Database            DatabaseConfig        `mapstructure:"database"`
	Redis               RedisConfig           `mapstructure:"redis"`
	Kafka               KafkaConfig           `mapstructure:"kafka"`
	JWT                 JWTConfig             `mapstructure:"jwt"`
	Log                 LogConfig             `mapstructure:"log"`
	Minio               MinioConfig           `mapstructure:"minio"`
	RustFS              RustFSConfig          `mapstructure:"rustfs"`
	Storage             StorageConfig         `mapstructure:"storage"`
	GRPC                GRPCClientConfig      `mapstructure:"grpc"`
	GRPCServer          GRPCServerConfig      `mapstructure:"grpc_server"`
	ServiceRegistry     ServiceRegistryConfig `mapstructure:"service_registry"`
	GRPCServiceRegistry ServiceRegistryConfig `mapstructure:"grpc_service_registry"`
	Dependencies        DependenciesConfig    `mapstructure:"dependencies"`
	Public              PublicConfig          `mapstructure:"public"`
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

// MinioConfig MinIO配置
type MinioConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
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

type StorageConfig struct {
	Provider string `mapstructure:"provider"`
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

// GRPCClientConfig gRPC客户端配置
type GRPCClientConfig struct {
	Timeout        time.Duration `mapstructure:"timeout"`
	MaxRecvMsgSize int           `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int           `mapstructure:"max_send_msg_size"`
	RetryTimes     int           `mapstructure:"retry_times"`
}

// ServiceRegistryConfig 服务注册配置
type ServiceRegistryConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	ServiceName     string        `mapstructure:"service_name"`
	ServiceID       string        `mapstructure:"service_id"`
	RegisterHost    string        `mapstructure:"register_host"`
	TTL             time.Duration `mapstructure:"ttl"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
}

// GRPCServerConfig describes the gRPC server bind endpoint.
type GRPCServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DependenciesConfig 依赖服务配置
type DependenciesConfig struct {
	UserService      UserServiceConfig      `mapstructure:"user_service"`
	TranscodeService TranscodeServiceConfig `mapstructure:"transcode_service"`
	VideoService     VideoServiceConfig     `mapstructure:"video_service"`
}

// UserServiceConfig 用户服务配置（支持地址或host+port）
type UserServiceConfig struct {
	ServiceName string        `mapstructure:"service_name"`
	Address     string        `mapstructure:"address"`
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

// TranscodeServiceConfig 转码服务配置（支持地址或host+port）
type TranscodeServiceConfig struct {
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

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 保持向后兼容：默认开启服务注册，明确配置可关闭
	viper.SetDefault("service_registry.enabled", true)
	viper.SetDefault("grpc_service_registry.enabled", true)
	viper.SetDefault("dependencies.user_service.service_name", "user-service")
	viper.SetDefault("dependencies.transcode_service.service_name", "transcode-service")
	viper.SetDefault("dependencies.video_service.service_name", "video-service")
	viper.SetDefault("storage.provider", "rustfs")
	viper.SetDefault("grpc_server.host", "0.0.0.0")
	// Kafka 默认
	viper.SetDefault("kafka.enabled", true)
	viper.SetDefault("kafka.client_id", "upload-service")
	viper.SetDefault("kafka.group_id", "upload-service-group")
	viper.SetDefault("kafka.bootstrap_servers", []string{"localhost:29092"})

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

	normalize(&config)

	return &config, nil
}

// normalize 填充默认值
func normalize(c *Config) {
	if c.ServiceRegistry.TTL == 0 {
		c.ServiceRegistry.TTL = 30 * time.Second
	}
	if c.ServiceRegistry.RefreshInterval == 0 {
		c.ServiceRegistry.RefreshInterval = 10 * time.Second
	}
	if c.GRPCServiceRegistry.TTL == 0 {
		c.GRPCServiceRegistry.TTL = 30 * time.Second
	}
	if c.GRPCServiceRegistry.RefreshInterval == 0 {
		c.GRPCServiceRegistry.RefreshInterval = 10 * time.Second
	}
	if c.GRPCServiceRegistry.RegisterHost == "" {
		c.GRPCServiceRegistry.RegisterHost = c.GRPCServer.Host
	}
	if c.GRPCServer.Host == "" {
		c.GRPCServer.Host = "0.0.0.0"
	}
	if c.Dependencies.UserService.Port == 0 {
		c.Dependencies.UserService.Port = 9091
	}
	if c.Dependencies.TranscodeService.Port == 0 {
		c.Dependencies.TranscodeService.Port = 9092
	}
	if c.Dependencies.VideoService.Port == 0 {
		c.Dependencies.VideoService.Port = 9093
	}
	if len(c.Kafka.BootstrapServers) == 0 {
		c.Kafka.BootstrapServers = []string{"localhost:29092"}
	}
	if c.Kafka.ClientID == "" {
		c.Kafka.ClientID = "upload-service"
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
	BootstrapServers []string `mapstructure:"bootstrap_servers"`
	ClientID         string   `mapstructure:"client_id"`
	GroupID          string   `mapstructure:"group_id"`
	Enabled          bool     `mapstructure:"enabled"`
}
