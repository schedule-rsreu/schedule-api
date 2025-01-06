package handlers

import (
	"net/http"

	v1 "github.com/schedule-rsreu/schedule-api/internal/http/handlers/v1"
	"github.com/schedule-rsreu/schedule-api/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/schedule-rsreu/schedule-api/docs"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// NewRouter
//
// @title           Schedule API
// @version         2.0
// @description     API for RSREU schedule.
// @externalDocs.description  GitHub
// @externalDocs.url          https://github.com/schedule-rsreu/schedule-api
func NewRouter(e *echo.Echo, scheduleService *services.ScheduleService) {
	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})

	e.GET("/docs/*", func(c echo.Context) error {
		if c.Request().URL.Path == "/docs/" {
			return c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
		}

		baseURL := c.Request().Host
		docs.SwaggerInfo.Host = baseURL

		return echoSwagger.WrapHandler(c)
	})

	v1.NewRouter(e.Group("/api/v1"), scheduleService)
}
