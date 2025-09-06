package db

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zetatez/assistant/internal/config"
	"github.com/zetatez/assistant/internal/models"
)

func New(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
	db, err := gorm.Open(mysql.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, err
	}

	// AutoMigrate demo models
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, err
	}

	log.Println("database connected and migrated")
	return db, nil
}
