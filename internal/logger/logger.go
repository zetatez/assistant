package logger

import (
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
	once   sync.Once
)

func GetLogger() *logrus.Logger { return logger }

func InitLogger(arg LoggerConfig) {
	once.Do(func() {
		logger = NewLogger(arg)
	})
	log.Println("✅ init log success !")
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
	JSONFormat bool   `mapstructure:"json_format"`
	Console    bool   `mapstructure:"console"`
}

func NewLogger(cfg LoggerConfig) *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if cfg.JSONFormat {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     cfg.Console,
		})
	}

	var outputs []io.Writer

	if cfg.Console {
		outputs = append(outputs, os.Stdout)
	}

	if cfg.Filename != "" {
		lj := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
		}
		outputs = append(outputs, lj)
	}

	if len(outputs) == 0 {
		outputs = append(outputs, os.Stdout)
	}

	logger.SetOutput(io.MultiWriter(outputs...))
	return logger
}
