// @title Assistant API
// @version 1.0
// @description 示例项目 API 文档
// @termsOfService http://example.com/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
package main

import (
	"log"

	"assistant/internal/config"
	"assistant/internal/db"
	"assistant/internal/module"
	"assistant/internal/modules/health"
	"assistant/internal/modules/user"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	cfg := config.Load()

	db.Init(cfg.DSN())

	r := gin.Default()

	modules := []module.Module{
		health.NewHealthModule(),
		user.NewUserModule(),
	}

	for _, m := range modules {
		log.Printf("migrating module: %s", m.Name())
		if err := m.Migrate(); err != nil {
			log.Fatalf("migration failed for %s: %v", m.Name(), err)
		}

		log.Printf("registering module: %s", m.Name())
		m.Register(r)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Server running at :8080")
	r.Run(":8080")
}
