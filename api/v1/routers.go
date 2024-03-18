package v1

import (
	"github.com/VinGP/schedule-api/docs"
	_ "github.com/VinGP/schedule-api/docs"
	"github.com/VinGP/schedule-api/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"net/http"
)

// NewRouter -.
// @title 		Расписание РГРТУ
// @description API для расписания РГРТУ
// @version     1.0
// @host        localhost:8081
// @BasePath    /api/v1
func NewRouter(handler *gin.Engine, s services.ScheduleService) {
	// Swagger
	//swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	//handler.GET("/swagger/*any", swaggerHandler)

	handler.GET("/swagger/*any", func(context *gin.Context) {
		docs.SwaggerInfo.Host = context.Request.Host
		ginSwagger.CustomWrapHandler(&ginSwagger.Config{URL: "/swagger/doc.json"}, swaggerFiles.Handler)(context)
	})

	// K8s probe
	handler.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Routers
	h := handler.Group("/api/v1")
	{
		newScheduleRoutes(h, s)
	}
}
