package prometheus_test

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	prom "github.com/prometheus/client_golang/prometheus"
	appmw "github.com/schedule-rsreu/schedule-api/internal/app"
	mw "github.com/schedule-rsreu/schedule-api/internal/http/middleware/prometheus"
)

func TestRequestMetricsKeepRouteFamilyAndExactCode(t *testing.T) {
	registry := prom.NewRegistry()
	previousRegisterer := prom.DefaultRegisterer
	prom.DefaultRegisterer = registry
	t.Cleanup(func() { prom.DefaultRegisterer = previousRegisterer })

	app := echo.New()
	app.Use(mw.NewPatternMiddleware("metrics-test"))
	app.Use(echo.WrapMiddleware(appmw.Heartbeat("/ping")))
	app.GET("/items/:id", func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) })
	app.POST("/items/:id", func(c *echo.Context) error { return c.NoContent(http.StatusCreated) })
	app.GET("/denied", func(c *echo.Context) error { return echo.NewHTTPError(http.StatusUnauthorized, "denied") })
	app.GET("/failed", func(c *echo.Context) error { return echo.NewHTTPError(http.StatusInternalServerError, "failed") })
	app.GET("/empty", func(c *echo.Context) error { return nil })

	cases := []struct {
		method, path, route, family, code string
		status                            int
	}{
		{http.MethodGet, "/items/private-id", "/items/:id", "2xx", "204", http.StatusNoContent},
		{http.MethodPost, "/items/another-id", "/items/:id", "2xx", "201", http.StatusCreated},
		{http.MethodGet, "/denied", "/denied", "4xx", "401", http.StatusUnauthorized},
		{http.MethodGet, "/failed", "/failed", "5xx", "500", http.StatusInternalServerError},
		{http.MethodGet, "/empty", "/empty", "2xx", "200", http.StatusOK},
		{http.MethodGet, "/ping", "/ping", "2xx", "200", http.StatusOK},
		{http.MethodGet, "/not-found/private-id", "unmatched", "4xx", "404", http.StatusNotFound},
	}
	expected := map[string]float64{}
	for _, test := range cases {
		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, http.NoBody))
		if recorder.Code != test.status {
			t.Fatalf("%s %s status=%d, want %d", test.method, test.path, recorder.Code, test.status)
		}
		expected[test.method+" "+test.route+" "+test.family+" "+test.code]++
	}
	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	actual := map[string]float64{}
	for _, family := range families {
		if family.GetName() != "http_requests_total" {
			continue
		}
		for _, metric := range family.Metric {
			labels := map[string]string{}
			for _, label := range metric.Label {
				labels[label.GetName()] = label.GetValue()
			}
			actual[labels["method"]+" "+labels["handler"]+" "+labels["status"]+" "+labels["code"]] += metric.GetCounter().GetValue()
		}
	}
	if !maps.Equal(actual, expected) {
		t.Fatalf("request counters=%v, want %v", actual, expected)
	}
}
