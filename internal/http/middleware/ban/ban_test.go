package ban_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/schedule-rsreu/schedule-api/internal/http/middleware/ban"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		ip         string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"lets allowed IP through", "192.0.2.1", "/api/v1/schedule/day", http.StatusNoContent, ""},
		{"forbids group schedule for banned IP", "203.0.113.7", "/api/v1/schedule/groups/344", http.StatusForbidden, "Access denied"},
		{"forbids other requests from banned IP", "203.0.113.7", "/api/v1/schedule/day", http.StatusForbidden, "Access denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			logger := zerolog.Nop()
			e.Use(ban.New([]string{"203.0.113.7"}, &logger))
			e.GET("/api/v1/schedule/groups/:group", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })
			e.GET("/api/v1/schedule/day", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set(echo.HeaderXRealIP, tt.ip)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if rec.Body.String() != tt.wantBody {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusForbidden && !strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), "text/plain") {
				t.Fatalf("Content-Type = %q, want text/plain", rec.Header().Get(echo.HeaderContentType))
			}
		})
	}
}

func TestMiddlewareLogsBannedIP(t *testing.T) {
	var output bytes.Buffer
	logger := zerolog.New(&output)
	e := echo.New()
	e.Use(ban.New([]string{"203.0.113.7"}, &logger))
	e.GET("/api/v1/schedule/day", func(c echo.Context) error { return c.NoContent(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/schedule/day", nil)
	req.Header.Set(echo.HeaderXRealIP, "203.0.113.7")
	e.ServeHTTP(httptest.NewRecorder(), req)

	log := output.String()
	for _, field := range []string{`"level":"warn"`, `"ip":"203.0.113.7"`, `"method":"GET"`, `"path":"/api/v1/schedule/day"`, `"message":"IP address blocked"`} {
		if !strings.Contains(log, field) {
			t.Fatalf("log = %q, want field %s", log, field)
		}
	}
}
