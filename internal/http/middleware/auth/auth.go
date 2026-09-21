package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/schedule-rsreu/schedule-api/pkg/auth/jwt"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

func New(telegramBotToken string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var authType, authData string
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				cookie, err := c.Cookie("access_token")
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header or cookie")
				}
				authType, authData = "Bearer", cookie.Value
			} else {
				var ok bool
				authType, authData, ok = strings.Cut(authHeader, " ")
				if !ok || authData == "" {
					return echo.NewHTTPError(http.StatusUnauthorized, "authorization header is invalid")
				}
			}

			switch {
			case strings.EqualFold(authType, "tma"):
				if err := initdata.Validate(authData, telegramBotToken, 7*24*time.Hour); err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
				}
			case strings.EqualFold(authType, "Bearer"):
				claims, err := jwt.ParseJWT(authData, []byte(telegramBotToken))
				if err != nil || claims.Type != jwt.AccessToken {
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
				}
			default:
				return echo.NewHTTPError(http.StatusUnauthorized, "unknown authorization type")
			}

			return next(c)
		}
	}
}
