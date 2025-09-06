package config

import (
	"assistant/internal/db"
	"assistant/internal/logger"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	config *Config
	once   sync.Once
)

func GetConfig() *Config { return config }

func InitConfig() {
	once.Do(func() {
		var err error
		config, err = LoadConfig()
		if err != nil {
			log.Fatalf("❌ init config failed: %v", err)
		}
	})
	log.Println("✅ init config success !")
}

type Config struct {
	App struct {
		Name string `mapstructure:"name"`
		Env  string `mapstructure:"env"`
		Addr string `mapstructure:"addr"`
	} `mapstructure:"app"`
	Log logger.LoggerConfig `mapstructure:"log"`
	DB  db.MySQLConfig      `mapstructure:"db"`
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
