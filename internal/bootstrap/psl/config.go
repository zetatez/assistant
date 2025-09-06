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
	App      AppConfig        `mapstructure:"app"`
	DB       xdb.DBPoolConfig `mapstructure:"db"`
	Dislock  DislockConfig    `mapstructure:"dislock"`
	Log      xlog.LogConfig   `mapstructure:"log"`
	LLM      LLMConfig        `mapstructure:"llm"`
	Channels ChannelsConfig   `mapstructure:"channels"`
	Tars     TarsConfig       `mapstructure:"tars"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int64  `mapstructure:"port"`
	Root struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Email    string `mapstructure:"email"`
	} `mapstructure:"root"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Feishu FeishuConfig `mapstructure:"feishu"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expiry int    `mapstructure:"expiry_hours"`
}

type FeishuConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

type DislockConfig struct {
	DefaultTTL int `mapstructure:"default_ttl"`
	MaxTTL     int `mapstructure:"max_ttl"`
}

type ChannelConfig struct {
	Provider string `mapstructure:"provider"`
}

type ChannelsConfig struct {
	Feishu FeishuConfig `mapstructure:"feishu"`
}

type TarsConfig struct {
	Enabled         bool           `mapstructure:"enabled"`
	LLMTemperature  float32        `mapstructure:"llm_temperature"`
	ChannelProvider string         `mapstructure:"channel_provider"`
	Channels        ChannelsConfig `mapstructure:"channels"`
	Persona         PersonaConfig  `mapstructure:"persona"`
	Memory          MemoryConfig   `mapstructure:"memory"`
}

type MemoryConfig struct {
	MaxHistory int `mapstructure:"max_history"`
	MemoryTTL  int `mapstructure:"memory_ttl_minutes"`
}

type PersonaConfig struct {
	HumorLevel   int `mapstructure:"humor_level"`
	HonestyLevel int `mapstructure:"honesty_level"`
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
	resolveEnvPlaceholder(&cfg.LLM.APIKey)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return cfg, nil
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
	if c.DB.MaxOpenConns <= 0 {
		c.DB.MaxOpenConns = 20
	}
	if c.DB.MaxIdleConns <= 0 {
		c.DB.MaxIdleConns = 10
	}
	if c.Dislock.DefaultTTL <= 0 {
		c.Dislock.DefaultTTL = 30
	}
	if c.Dislock.MaxTTL <= 0 {
		c.Dislock.MaxTTL = 300
	}
	if c.Dislock.MaxTTL < c.Dislock.DefaultTTL {
		return fmt.Errorf("dislock.max_ttl (%d) must be >= dislock.default_ttl (%d)", c.Dislock.MaxTTL, c.Dislock.DefaultTTL)
	}
	if c.App.JWT.Secret == "" {
		return fmt.Errorf("app.jwt.secret is required")
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
		c.Tars.Memory.MaxHistory = 20
	}
	if c.Tars.Memory.MemoryTTL <= 0 {
		c.Tars.Memory.MemoryTTL = 60
	}
	if c.Tars.LLMTemperature <= 0 {
		c.Tars.LLMTemperature = 0.7
	}
	if c.Tars.ChannelProvider == "" {
		c.Tars.ChannelProvider = "feishu"
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
	return nil
}

var envPlaceholder = regexp.MustCompile(`\$\{(\w+)\}`)

func resolveEnvPlaceholder(val *string) {
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

func (c *LLMConfig) AfterMerge() {
	c.APIKey = ResolveEnvPlaceholderStr(c.APIKey)
}
