package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置，与其他服务保持结构一致，便于共享启动逻辑。
type Config struct {
	Server          ServerConfig          `mapstructure:"server"`
	Database        DatabaseConfig        `mapstructure:"database"`
	Redis           RedisConfig           `mapstructure:"redis"`
	Kafka           KafkaConfig           `mapstructure:"kafka"`
	Log             LogConfig             `mapstructure:"log"`
	Minio           MinioConfig           `mapstructure:"minio"`
	GRPC            GRPCConfig            `mapstructure:"grpc"`
	ServiceRegistry ServiceRegistryConfig `mapstructure:"service_registry"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

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

type MinioConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

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

type GRPCConfig struct {
	Port           int           `mapstructure:"port"`
	Network        string        `mapstructure:"network"`
	Timeout        time.Duration `mapstructure:"timeout"`
	MaxRecvMsgSize int           `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int           `mapstructure:"max_send_msg_size"`
}

type ServiceRegistryConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	ServiceName     string        `mapstructure:"service_name"`
	ServiceID       string        `mapstructure:"service_id"`
	RegisterHost    string        `mapstructure:"register_host"`
	TTL             time.Duration `mapstructure:"ttl"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	BootstrapServers []string `mapstructure:"bootstrap_servers"`
	ClientID         string   `mapstructure:"client_id"`
	GroupID          string   `mapstructure:"group_id"`
	Enabled          bool     `mapstructure:"enabled"`
}

// Load 加载配置文件。
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 默认开启服务注册以保持向后兼容，可通过配置关闭
	viper.SetDefault("service_registry.enabled", true)

	// Kafka 默认
	viper.SetDefault("kafka.enabled", true)
	viper.SetDefault("kafka.client_id", "notification-service")
	viper.SetDefault("kafka.group_id", "notification-service-group")
	viper.SetDefault("kafka.bootstrap_servers", []string{"localhost:29092"})

	// 设置环境变量前缀
	viper.SetEnvPrefix("GO_VIDEO")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	normalize(&cfg)
	return &cfg, nil
}

// normalize 补全默认值
func normalize(c *Config) {
	if c.ServiceRegistry.TTL == 0 {
		c.ServiceRegistry.TTL = 30 * time.Second
	}
	if c.ServiceRegistry.RefreshInterval == 0 {
		c.ServiceRegistry.RefreshInterval = 10 * time.Second
	}
}

// GetDSN 构建 MySQL DSN。
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

// GetRedisAddr 获取 Redis 地址。
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
