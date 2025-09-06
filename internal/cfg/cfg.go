package cfg

import (
	"assistant/pkg/xdb"
	"assistant/pkg/xlog"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	C    *Config
	once sync.Once
)

func InitCFG() {
	once.Do(func() {
		var err error
		C, err = LoadConfig()
		if err != nil {
			log.Fatalf("❌ init config failed: %v", err)
		}
	})
	log.Println("✅ init config success !")
}

type Config struct {
	App struct {
		Name string `mapstructure:"name"`
		Port int64  `mapstructure:"port"`
	} `mapstructure:"app"`
	Log xlog.LogConfig  `mapstructure:"log"`
	DB  xdb.MySQLConfig `mapstructure:"db"`
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
