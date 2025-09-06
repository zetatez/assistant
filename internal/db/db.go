package db

import (
	"assistant/internal/cfg"
	"assistant/internal/log"
	"assistant/pkg/xdb"
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db   *sql.DB
	once sync.Once
)

func GetDB() *sql.DB { return db }

func InitDB() {
	once.Do(func() {
		var err error
		db, err = xdb.NewMySQL(cfg.C.DB)
		if err != nil {
			log.Logger.Fatalf("❌ init db failed: %v", err)
		}
	})
	log.Logger.Infof("✅ init db success !")
}
