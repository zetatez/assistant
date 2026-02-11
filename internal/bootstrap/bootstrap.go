package bootstrap

import (
	"assistant/internal/app"
	"assistant/internal/bootstrap/psl"
)

func Init() {
	psl.InitConfig()

	psl.InitLog()

	psl.InitDB()

	psl.InitDisLocker()

	psl.MigrateDB()

	app.Run()
}
