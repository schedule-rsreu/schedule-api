package dwh

import (
	"strings"

	logger2 "github.com/schedule-rsreu/schedule-api/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/schedule-rsreu/schedule-api/pkg/auth/jwt"
)

func getBearerClaims(c echo.Context, bearerSecret string) (*jwt.Claims, bool) {
	authHeader := c.Request().Header.Get("authorization")

	if authHeader == "" {
		return nil, false
	}

	authParts := strings.Split(authHeader, " ")

	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return nil, false
	}

	claims, err := jwt.ParseJWT(authParts[1], []byte(bearerSecret))

	if err != nil {
		return nil, false
	}

	return claims, true
}

func New(dwhURL, dwhToken, bearerSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.Contains(c.Request().RequestURI, "/ping") || strings.Contains(c.Request().RequestURI, "/metrics") {
				return next(c)
			}
			logger := logger2.GetLoggerFromCtx(c)
			claims, ok := getBearerClaims(c, bearerSecret)
			logger.Info().Any("claims", claims).Bool("ok", ok).Msg("dwh middleware")
			return next(c)
		}
	}
}
