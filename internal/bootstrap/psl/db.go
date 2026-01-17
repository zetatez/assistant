package psl

import (
	"assistant/pkg/xdb"
	"context"
	"database/sql"
	"sync"
)

var (
	db     *sql.DB
	onceDB sync.Once
)

func GetDB() *sql.DB { return db }

func InitDB() {
	onceDB.Do(func() {
		var err error
		db, err = xdb.NewDBPool(context.Background(), GetConfig().DB)
		if err != nil {
			GetLogger().Fatalf("init db failed: %v", err)
		}
	})
	GetLogger().Infof("init db success")
}
