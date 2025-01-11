package logger

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const LoggerCtxKey = "logger"

func GetLoggerFromCtx(c echo.Context) *zerolog.Logger {
	logger, ok := c.Get(LoggerCtxKey).(*zerolog.Logger)

	if !ok {
		return &log.Logger
	}

	return logger
}
