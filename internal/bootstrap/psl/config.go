package psl

import (
	"assistant/pkg/xdb"
	"assistant/pkg/xlog"
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	config     *Config
	onceConfig sync.Once
)

func GetConfig() *Config { return config }

func InitConfig() {
	onceConfig.Do(
		func() {
			var err error
			config, err = LoadConfig()
			if err != nil {
				panic(err)
			}
		},
	)
}

type Config struct {
	App struct {
		Name string `mapstructure:"name"`
		Port int64  `mapstructure:"port"`
		Root struct {
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			Email    string `mapstructure:"email"`
		} `mapstructure:"root"`
	} `mapstructure:"app"`
	DB      xdb.DBPoolConfig `mapstructure:"db"`
	Dislock DislockConfig    `mapstructure:"dislock"`
	Log     xlog.LogConfig   `mapstructure:"log"`
}

type DislockConfig struct {
	DefaultTTL int `mapstructure:"default_ttl"`
	MaxTTL     int `mapstructure:"max_ttl"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read in config: %v", err)
	}
	var cfg *Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed unmarshal config: %v", err)
	}
	return cfg, nil
}
