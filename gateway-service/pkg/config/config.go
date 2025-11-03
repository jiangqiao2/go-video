package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the full gateway configuration.
type Config struct {
	Server   ServerConfig             `mapstructure:"server"`
	Log      LogConfig                `mapstructure:"log"`
	JWT      JWTConfig                `mapstructure:"jwt"`
	Services map[string]ServiceConfig `mapstructure:"services"`
	Routes   []RouteConfig            `mapstructure:"routes"`
}

// ServerConfig defines HTTP server runtime options.
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// LogConfig controls structured logging output.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// JWTConfig provides token validation settings.
type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	ExpireTime        time.Duration `mapstructure:"expire_time"`
	RefreshExpireTime time.Duration `mapstructure:"refresh_expire_time"`
}

// ServiceConfig declares an upstream service that the gateway can forward to.
type ServiceConfig struct {
	BaseURL       string        `mapstructure:"base_url"`
	Timeout       time.Duration `mapstructure:"timeout"`
	StripPrefix   string        `mapstructure:"strip_prefix"`
	PreserveHost  bool          `mapstructure:"preserve_host"`
	StaticHeaders map[string]string `mapstructure:"static_headers"`
}

// RouteConfig maps incoming path prefixes to upstream services.
type RouteConfig struct {
	Name          string            `mapstructure:"name"`
	PathPrefix    string            `mapstructure:"path_prefix"`
	Methods       []string          `mapstructure:"methods"`
	TargetService string            `mapstructure:"target_service"`
	AuthRequired  bool              `mapstructure:"auth_required"`
	StripPrefix   string            `mapstructure:"strip_prefix"`
	StaticHeaders map[string]string `mapstructure:"static_headers"`
}

// Load reads configuration from the provided file path.
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "configs/config.dev.yaml"
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	if ext := filepath.Ext(configPath); ext != "" {
		v.SetConfigType(strings.TrimPrefix(ext, "."))
	}
	v.SetEnvPrefix("GATEWAY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	normalizeConfig(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate performs basic sanity checks on the configuration.
func (cfg *Config) Validate() error {
	if cfg.Server.Port <= 0 {
		return fmt.Errorf("server.port must be > 0")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret must be provided")
	}
	for _, route := range cfg.Routes {
		if route.PathPrefix == "" {
			return fmt.Errorf("route %q has empty path_prefix", route.Name)
		}
		if route.TargetService == "" {
			return fmt.Errorf("route %q missing target_service", route.Name)
		}
		if _, ok := cfg.Services[route.TargetService]; !ok {
			return fmt.Errorf("route %q references undefined service %q", route.Name, route.TargetService)
		}
	}
	return nil
}

func normalizeConfig(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "release"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}

	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "json"
	}
	if cfg.Log.Output == "" {
		cfg.Log.Output = "stdout"
	}

	for svcName, svc := range cfg.Services {
		if svc.Timeout == 0 {
			svc.Timeout = 15 * time.Second
		}
		if svc.StripPrefix == "" {
			svc.StripPrefix = ""
		}
		cfg.Services[svcName] = svc
	}

	for i := range cfg.Routes {
		cfg.Routes[i].PathPrefix = ensureLeadingSlash(cfg.Routes[i].PathPrefix)
		if cfg.Routes[i].StripPrefix == "" && cfg.Services != nil {
			cfg.Routes[i].StripPrefix = cfg.Services[cfg.Routes[i].TargetService].StripPrefix
		}
		if len(cfg.Routes[i].Methods) > 0 {
			for j := range cfg.Routes[i].Methods {
				cfg.Routes[i].Methods[j] = strings.ToUpper(cfg.Routes[i].Methods[j])
			}
		}
	}
}

func ensureLeadingSlash(path string) string {
	if path == "" || path[0] == '/' {
		return path
	}
	return "/" + path
}
