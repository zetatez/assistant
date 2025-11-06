package log

import (
	"sync"

	"assistant/internal/cfg"
	"assistant/pkg/xlog"

	"github.com/sirupsen/logrus"
)

var (
	Logger *logrus.Logger
	once   sync.Once
)

func InitLog() {
	once.Do(func() {
		Logger = xlog.NewLogger(cfg.C.Log)
	})
	Logger.Println("✅ init logger success !")
}
