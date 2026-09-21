package ban

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func New(ips []string, logger *zerolog.Logger) echo.MiddlewareFunc {
	banned := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		if ip = strings.TrimSpace(ip); ip != "" {
			banned[ip] = struct{}{}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if _, ok := banned[ip]; !ok {
				return next(c)
			}
			logger.Warn().Str("ip", ip).Str("method", c.Request().Method).Str("path", c.Request().URL.Path).Msg("IP address blocked")
			return c.String(http.StatusForbidden, "Access denied")
		}
	}
}
