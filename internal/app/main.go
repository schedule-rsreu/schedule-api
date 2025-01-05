package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/schedule-rsreu/schedule-api/internal/repo"
	"github.com/schedule-rsreu/schedule-api/internal/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/schedule-rsreu/schedule-api/internal/http/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/schedule-rsreu/schedule-api/config"
)

const contextTimeout = 5 * time.Second

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func Run(cfg *config.Config) {
	logger := zerolog.New(os.Stdout)

	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	setupLogger(e, &logger)

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.GetMongoURI()))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		logger.Error().Err(err).Msg("MongoDB ping failed")
		return
	}

	handlers.NewRouter(e, services.NewScheduleService(repo.New(client)))

	go func() {
		logger.Debug().Msg("Open http://" + net.JoinHostPort("localhost", cfg.Port) + "/docs")
		err := e.Start(net.JoinHostPort(cfg.Host, cfg.Port))
		if err != nil {
			logger.Error().Err(err).Msg("app - Run - httpServer.Start")
			return
		}
	}()

	logger.Info().Msg("Server started")

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	s := <-interrupt
	logger.Info().Str("signal", s.String()).Msg("app - Run - signal.Notify")

	ctx, cancel = context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("app - Run - httpServer.Shutdown")
	}
}

func setupLogger(e *echo.Echo, logger *zerolog.Logger) {
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", logger)
			return next(c)
		}
	})

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:          true,
		LogStatus:       true,
		LogLatency:      true,
		LogResponseSize: true,
		LogMethod:       true,
		LogUserAgent:    true,
		LogRequestID:    true,
		LogError:        true,

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

			logger.Info().
				Time("time", v.StartTime).
				Str("component", "middleware").
				Str("path", v.URI).
				Str("method", v.Method).
				Str("remote_addr", v.RemoteIP).
				Str("user_agent", v.UserAgent).
				Str("request_id", v.RequestID).
				Int("status", status).
				Int64("bytes", v.ResponseSize).
				Str("duration", v.Latency.String()).
				Err(v.Error).
				Msg("request completed")

			return nil
		},
	}))
}
