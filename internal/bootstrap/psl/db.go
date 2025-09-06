package psl

import (
	"assistant/pkg/xdb"
	"context"
	"database/sql"
	"fmt"
	"sync"
)

var (
	db     *sql.DB
	onceDB sync.Once
)

func GetDB() *sql.DB { return db }

func InitDB(ctx context.Context) error {
	var initErr error
	onceDB.Do(func() {
		var err error
		db, err = xdb.NewDBPool(ctx, GetConfig().DB)
		if err != nil {
			initErr = fmt.Errorf("new db pool: %w", err)
			return
		}
	})
	return initErr
}
