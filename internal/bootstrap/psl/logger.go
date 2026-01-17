package psl

import (
	"sync"

	"assistant/pkg/xlog"

	"github.com/sirupsen/logrus"
)

var (
	logger  *logrus.Logger
	onceLog sync.Once
)

func GetLogger() *logrus.Logger { return logger }

func InitLog() {
	onceLog.Do(func() {
		logger = xlog.NewLogger(GetConfig().Log)
	})
	logger.Println("init log success")
}
