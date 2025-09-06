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
	"assistant/internal/app"
	"assistant/internal/bootstrap"
)

func main() {
	bootstrap.Init()
	app.Run()
}
