package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"

	"github.com/schedule-rsreu/schedule-api/pkg/postgres"

	"github.com/schedule-rsreu/schedule-api/internal/http/middleware/dwh"

	"github.com/labstack/gommon/color"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	mwp "github.com/schedule-rsreu/schedule-api/internal/http/middleware/prometheus"

	"github.com/go-playground/validator/v10"

	"github.com/schedule-rsreu/schedule-api/internal/http/handlers"
	"github.com/schedule-rsreu/schedule-api/internal/repo"
	"github.com/schedule-rsreu/schedule-api/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/schedule-rsreu/schedule-api/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const contextTimeout = 5 * time.Second
const serviceName = "schedule-api"

const productionUrl = "https://schedule-rsreu.ru"
const banner = `
   _____      __             __      __        ___    ____  ____
  / ___/_____/ /_  ___  ____/ /_  __/ /__     /   |  / __ \/  _/
  \__ \/ ___/ __ \/ _ \/ __  / / / / / _ \   / /| | / /_/ // /  
 ___/ / /__/ / / /  __/ /_/ / /_/ / /  __/  / ___ |/ ____// /   
/____/\___/_/ /_/\___/\__,_/\__,_/_/\___/  /_/  |_/_/   /___/   %s
API for RSREU schedule
___________________________________________

⇨ server started on: %s
⇨ docs: %s
`

// initTracer возвращает TracerProvider в зависимости от окружения.
func initTracer(ctx context.Context, otlEndpoint, env, version string, logger *zerolog.Logger) (*sdktrace.TracerProvider, error) {
	if env == "local" {
		logger.Info().Msg("⚙️  Using no-op tracer (local mode)")
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)
		otel.SetTracerProvider(tp)
		return tp, nil
	}

	// иначе — подключаем реальный экспортер
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otlEndpoint),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("deployment.environment", env),
			attribute.String("service.version", version),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	logger.Info().Msgf("🚀 Tracing enabled for environment=%s", env)
	return tp, nil
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func Heartbeat(endpoint string) func(http.Handler) http.Handler {
	f := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if (r.Method == http.MethodGet || r.Method == http.MethodHead) && strings.EqualFold(r.URL.Path, endpoint) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("."))

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return f
}

func printBanner(version, url string) {
	colorer := color.New()

	colorer.Printf(banner,
		colorer.Red("v"+version),
		colorer.Blue(url),
		colorer.BlueBg(url+"/docs/index.html"),
	)
}

func Run(cfg *config.Config) {
	logger := zerolog.New(os.Stdout)

	ctx := context.Background()

	tp, err := initTracer(ctx, cfg.OtlEndpoint, cfg.Environment, cfg.Version, &logger)
	if err != nil {
		logger.Error().Err(err).Msg("app - Run - initTracer")
		return
	}

	defer func(tp *sdktrace.TracerProvider, ctx context.Context) {
		err = tp.Shutdown(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("app - Run - tp.Shutdown")
		}
	}(tp, ctx)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	postgresDB, err := postgres.New(cfg.PostgresDSN)
	if err != nil {
		logger.Error().Err(err).Msg("Postgres connection failed")
		return
	}

	handlers.NewRouter(e, services.NewScheduleService(repo.NewScheduleRepo(postgresDB)))

	go func() {
		if cfg.Production {
			printBanner(cfg.Version, productionUrl)
		} else {
			printBanner(cfg.Version, "http://localhost:"+cfg.Port)
		}
		setupEcho(e, &logger)

		err := e.Start(net.JoinHostPort(cfg.Host, cfg.Port))
		if err != nil {
			logger.Error().Err(err).Msg("app - Run - httpServer.Start")
			return
		}
	}()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	logger.Info().Str("signal", s.String()).Msg("app - Run - signal.Notify")

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("app - Run - httpServer.Shutdown")
	}

	postgresDB.Close()
	logger.Info().Msg("app - Run - postgresDB.Close - exit")

	logger.Info().Msg("app - Run - exit")
}

func setupEcho(e *echo.Echo, logger *zerolog.Logger) {
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	setupLogger(e, logger)

	e.Use(mwp.NewPatternMiddleware("schedule_api"))
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.Use(echo.WrapMiddleware(Heartbeat("/ping")))

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(dwh.New("", "", "secret"))
}

func addTraceToLogMiddleware(defaultLogger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		tracer := otel.Tracer("schedule-api/handlers")

		return func(c echo.Context) error {
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)

			ctx := otel.GetTextMapPropagator().Extract(
				c.Request().Context(),
				propagation.HeaderCarrier(c.Request().Header),
			)

			// создаём span для запроса
			ctx, span := tracer.Start(ctx, c.Path(), trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()
			traceID := span.SpanContext().TraceID().String()
			c.Set("trace_id", traceID)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Response().Header().Set("X-Trace-ID", traceID)

			loggerFromCtx, ok := c.Get("logger").(*zerolog.Logger)
			if !ok {
				loggerFromCtx = defaultLogger
			}
			logger := loggerFromCtx.With().
				Str("trace_id", traceID).
				Str("request_id", reqID).
				Logger()
			c.Set("logger", &logger)

			err := next(c)

			status := c.Response().Status
			if err != nil {
				var he *echo.HTTPError
				if errors.As(err, &he) {
					status = he.Code
				} else {
					status = http.StatusInternalServerError
				}
				span.SetAttributes(attribute.String("error", err.Error()))
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}
			span.SetAttributes(attribute.Int("http.status_code", status))

			return err
		}
	}
}

func logRequestMiddleware(logger *zerolog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:          true,
		LogStatus:       true,
		LogLatency:      true,
		LogResponseSize: true,
		LogMethod:       true,
		LogUserAgent:    true,
		LogRequestID:    true,
		LogError:        true,
		LogRemoteIP:     true,

		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			status := v.Status
			if v.Error != nil {
				var errWithStatus *echo.HTTPError
				switch {
				case errors.As(v.Error, &errWithStatus):
					status = errWithStatus.Code
				default:
					status = http.StatusInternalServerError
				}
			}
			loggerFromCtx, ok := c.Get("logger").(*zerolog.Logger)

			if !ok {
				loggerFromCtx = logger
			}

			loggerFromCtx.Info().
				Time("time", v.StartTime).
				Str("component", "middleware").
				Str("path", v.URI).
				Str("method", v.Method).
				Str("remote_addr", v.RemoteIP).
				Str("user_agent", v.UserAgent).
				Int("status", status).
				Int64("bytes", v.ResponseSize).
				Str("duration", v.Latency.String()).
				Err(v.Error).
				Msg("request completed")

			return nil
		},
	})
}

func setupLogger(e *echo.Echo, logger *zerolog.Logger) {
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", logger)
			return next(c)
		}
	})

	e.Use(addTraceToLogMiddleware(logger))

	e.Use(logRequestMiddleware(logger))
}
