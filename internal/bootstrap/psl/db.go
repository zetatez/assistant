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
		cfg := GetConfig().DB
		poolCfg := xdb.DBPoolConfig{
			DriverName:      cfg.Driver,
			DSN:             cfg.DSN,
			MaxOpenConns:    cfg.Pool.MaxOpenConns,
			MaxIdleConns:    cfg.Pool.MaxIdleConns,
			ConnMaxLifetime: cfg.Pool.ConnMaxLifetime,
			ConnMaxIdletime: cfg.Pool.ConnMaxIdletime,
		}
		db, err = xdb.NewDBPool(ctx, poolCfg)
		if err != nil {
			initErr = fmt.Errorf("new db pool: %w", err)
			return
		}
	})
	return initErr
}
