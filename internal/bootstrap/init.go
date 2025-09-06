package bootstrap

import (
	"assistant/internal/config"
	"assistant/internal/db"
	"assistant/internal/logger"
	"assistant/internal/migration"
	"log"
)

func Init() {
	// init app config
	config.InitConfig()

	// init log
	logger.InitLogger(config.GetConfig().Log)

	// init db
	db.InitDB(logger.GetLogger(), config.GetConfig().DB)

	// migration
	migration.Migrate(logger.GetLogger(), db.GetDB())

	log.Println("✅ app is initialized !")
}
