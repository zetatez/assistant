package bootstrap

import (
	"assistant/internal/app"
	"assistant/internal/cfg"
	"assistant/internal/db"
	"assistant/internal/log"
	"assistant/internal/migration"
)

func Init() {
	cfg.InitCFG()

	log.InitLog()

	db.InitDB()

	migration.Migrate()

	log.Logger.Info("✅ app is initialized !")

	app.Run()
}
