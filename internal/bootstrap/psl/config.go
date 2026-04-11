package psl

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"assistant/pkg/xdb"
	"assistant/pkg/xlog"

	"github.com/spf13/viper"
)

var (
	config     *Config
	onceConfig sync.Once
)

func GetConfig() *Config { return config }

func InitConfig() error {
	var initErr error
	onceConfig.Do(func() {
		var err error
		config, err = LoadConfig()
		if err != nil {
			initErr = fmt.Errorf("load config failed: %w", err)
			return
		}
	})
	return initErr
}

type Config struct {
	App     AppConfig      `mapstructure:"app"`
	Auth    AuthConfig     `mapstructure:"auth"`
	DB      DBConfig       `mapstructure:"db"`
	Log     xlog.LogConfig `mapstructure:"log"`
	LLM     LLMConfig      `mapstructure:"llm"`
	Channel ChannelConfig  `mapstructure:"channel"`
	Tars    TarsConfig     `mapstructure:"tars"`
	Monitor MonitorConfig  `mapstructure:"monitor"`
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	Port      int    `mapstructure:"port"`
	Interface string `mapstructure:"interface"`
}

type AuthConfig struct {
	Root struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Email    string `mapstructure:"email"`
	} `mapstructure:"root"`
	JWT struct {
		Secret string `mapstructure:"secret"`
		Expiry int    `mapstructure:"expiry_hours"`
	} `mapstructure:"jwt"`
}

type DBConfig struct {
	Driver string           `mapstructure:"driver"`
	DSN    string           `mapstructure:"dsn"`
	Pool   xdb.DBPoolConfig `mapstructure:"pool"`
}

type LLMConfig struct {
	Provider    string            `mapstructure:"provider"`
	APIKey      string            `mapstructure:"api_key"`
	BaseURL     string            `mapstructure:"base_url"`
	Model       string            `mapstructure:"model"`
	Extra       map[string]string `mapstructure:"extra"`
	Timeout     int               `mapstructure:"timeout"`
	MaxTokens   int               `mapstructure:"max_tokens"`
	Temperature float32           `mapstructure:"temperature"`
}

type ChannelConfig struct {
	Provider string       `mapstructure:"provider"`
	Feishu   FeishuConfig `mapstructure:"feishu"`
}

type FeishuConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

type TarsConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	LLMTemperature float32       `mapstructure:"llm_temperature"`
	Persona        PersonaConfig `mapstructure:"persona"`
	Memory         MemoryConfig  `mapstructure:"memory"`
	Wiki           WikiConfig    `mapstructure:"wiki"`
}

type PersonaConfig struct {
	HumorLevel   int `mapstructure:"humor_level"`
	HonestyLevel int `mapstructure:"honesty_level"`
}

type MemoryConfig struct {
	MaxHistory int `mapstructure:"max_history"`
	TTLMinutes int `mapstructure:"ttl_minutes"`
}

type WikiConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Dir     string `mapstructure:"dir"`
}

type MonitorConfig struct {
	Tracing TracingConfig `mapstructure:"tracing"`
	Metrics MetricsConfig `mapstructure:"metrics"`
}

type TracingConfig struct {
	Enabled    bool    `mapstructure:"enabled"`
	SampleRate float32 `mapstructure:"sample_rate"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read in config: %v", err)
	}
	var cfg *Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed unmarshal config: %v", err)
	}
	cfg.resolveEnv()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return cfg, nil
}

func (c *Config) resolveEnv() {
	resolveEnv(&c.LLM.APIKey)
}

func (c *Config) Validate() error {
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return fmt.Errorf("app.port must be between 1 and 65535, got %d", c.App.Port)
	}
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if c.DB.DSN == "" {
		return fmt.Errorf("db.dsn is required")
	}
	if c.DB.Pool.MaxOpenConns <= 0 {
		c.DB.Pool.MaxOpenConns = 20
	}
	if c.DB.Pool.MaxIdleConns <= 0 {
		c.DB.Pool.MaxIdleConns = 10
	}
	if c.LLM.Provider == "" {
		return fmt.Errorf("llm.provider is required")
	}
	if c.LLM.Timeout <= 0 {
		c.LLM.Timeout = 60
	}
	if c.LLM.MaxTokens <= 0 {
		c.LLM.MaxTokens = 4096
	}
	if c.Tars.Memory.MaxHistory <= 0 {
		c.Tars.Memory.MaxHistory = 64
	}
	if c.Tars.Memory.TTLMinutes <= 0 {
		c.Tars.Memory.TTLMinutes = 60
	}
	if c.Tars.LLMTemperature <= 0 {
		c.Tars.LLMTemperature = 0.7
	}
	if c.Tars.Persona.HumorLevel < 0 {
		c.Tars.Persona.HumorLevel = 0
	}
	if c.Tars.Persona.HumorLevel > 100 {
		c.Tars.Persona.HumorLevel = 100
	}
	if c.Tars.Persona.HonestyLevel < 0 {
		c.Tars.Persona.HonestyLevel = 0
	}
	if c.Tars.Persona.HonestyLevel > 100 {
		c.Tars.Persona.HonestyLevel = 100
	}
	if c.Monitor.Tracing.SampleRate <= 0 {
		c.Monitor.Tracing.SampleRate = 1.0
	}
	if c.Monitor.Tracing.SampleRate > 1 {
		c.Monitor.Tracing.SampleRate = 1.0
	}
	if c.Monitor.Metrics.Path == "" {
		c.Monitor.Metrics.Path = "/metrics"
	}
	return nil
}

var envPlaceholder = regexp.MustCompile(`\$\{(\w+)\}`)

func resolveEnv(val *string) {
	if val == nil || *val == "" {
		return
	}
	matches := envPlaceholder.FindStringSubmatch(*val)
	if len(matches) >= 2 {
		*val = os.Getenv(matches[1])
	}
}

func ResolveEnvPlaceholderStr(val string) string {
	if val == "" {
		return ""
	}
	matches := envPlaceholder.FindStringSubmatch(val)
	if len(matches) >= 2 {
		return os.Getenv(matches[1])
	}
	return val
}

func (c *LLMConfig) GetAPIKey() string {
	return ResolveEnvPlaceholderStr(c.APIKey)
}
