package prometheus

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	reqsName         = "http_requests_total"
	latencyHighrName = "http_request_duration_highr_seconds"
	latencyLowrName  = "http_request_duration_seconds"
)

// Middleware is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method, and HTTP path.
type Middleware struct {
	reqs         *prometheus.CounterVec
	latencyLowr  *prometheus.HistogramVec
	latencyHighr *prometheus.HistogramVec
}

// NewPatternMiddleware returns a new prometheus Middleware handler that groups requests by Echo route pattern.
func NewPatternMiddleware(name string) echo.MiddlewareFunc {
	var (
		latencyLowrBuckets  = []float64{0.1, 0.5, 1}
		latencyHighrBuckets = []float64{0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 7.5, 10, 30, 60} //nolint:lll // .
	)

	var m Middleware
	m.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path (with patterns).",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"handler", "method", "status"},
	)
	prometheus.MustRegister(m.reqs)

	m.latencyHighr = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyHighrName,
		Help:        "Latency with many buckets but no API-specific labels. For more accurate percentile calculations.",
		ConstLabels: prometheus.Labels{"service": name},
		Buckets:     latencyHighrBuckets,
	},
		[]string{},
	)
	prometheus.MustRegister(m.latencyHighr)

	m.latencyLowr = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyLowrName,
		Help:        "Latency with only few buckets by handler. For aggregation by handler.",
		ConstLabels: prometheus.Labels{"service": name},
		Buckets:     latencyLowrBuckets,
	},
		[]string{"handler", "method", "status"},
	)
	prometheus.MustRegister(m.latencyLowr)

	return m.patternHandler
}

func (m Middleware) patternHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Execute the handler
		err := next(c)

		// Capture the metrics after the handler executes
		statusCode := c.Response().Status

		if err != nil {
			var httpError *echo.HTTPError
			if errors.As(err, &httpError) {
				statusCode = httpError.Code
			}
			if statusCode == 0 || statusCode == http.StatusOK {
				statusCode = http.StatusInternalServerError
			}
		}

		const statusCodeDivisor = 100

		status := strconv.Itoa(statusCode/statusCodeDivisor) + "xx"
		routePattern := c.Path()
		routePattern = strings.ReplaceAll(routePattern, "/*", "")

		duration := time.Since(start).Seconds()

		// Update Prometheus metrics
		m.reqs.WithLabelValues(routePattern, c.Request().Method, status).Inc()
		m.latencyHighr.WithLabelValues().Observe(duration)
		m.latencyLowr.WithLabelValues(routePattern, c.Request().Method, status).Observe(duration)

		return err
	}
}
