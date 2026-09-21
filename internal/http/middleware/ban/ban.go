package ban

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func New(ips []string) echo.MiddlewareFunc {
	banned := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		if ip = strings.TrimSpace(ip); ip != "" {
			banned[ip] = struct{}{}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, ok := banned[c.RealIP()]; !ok {
				return next(c)
			}
			return c.String(http.StatusForbidden, "Access denied")
		}
	}
}
