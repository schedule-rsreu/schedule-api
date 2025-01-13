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

	"github.com/schedule-rsreu/schedule-api/pkg/mongodb"

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
)

const contextTimeout = 5 * time.Second

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

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	mongoClient, err := mongodb.NewMongoClient(cfg.GetMongoURI())

	if err != nil {
		logger.Error().Err(err).Msg("MongoDB ping failed")
		return
	}
	scheduleDB := mongodb.NewMongoDatabase(mongoClient, cfg.MongoDBName)

	handlers.NewRouter(e, services.NewScheduleService(repo.NewScheduleRepo(scheduleDB)))

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

	if err := mongoClient.Disconnect(ctx); err != nil {
		logger.Error().Err(err).Msg("app - Run - mongoClient.Disconnect")
		return
	}
	logger.Info().Msg("app - Run - mongoClient.Disconnect - exit")

	logger.Info().Msg("app - Run - exit")
}

func setupEcho(e *echo.Echo, logger *zerolog.Logger) {
	e.Use(mwp.NewPatternMiddleware("schedule_api"))
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.Use(echo.WrapMiddleware(Heartbeat("/ping")))

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	setupLogger(e, logger)
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
