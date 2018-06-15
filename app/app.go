package app

import (
	"net/http"

	"git.gogoair.com/bagws/lambdagateway/app/payment"
	"github.com/gobuffalo/envy"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

// NewApp creates and returns a new router
func NewApp() *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.BodyLimit("1M"))
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.GET("/healthcheck", healthHandler)
	config(e)

	// Register payment routes
	payment.Register(e)

	return e
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Everything is awesome!")
}

func config(e *echo.Echo) {
	e.Logger.SetLevel(logLevel(envy.Get("LOG_LEVEL", "")))
}

func logLevel(lvl string) log.Lvl {
	switch lvl {
	case "DEBUG":
		return log.DEBUG
	case "INFO":
		return log.INFO
	case "WARN":
		return log.WARN
	case "ERROR":
		return log.ERROR
	case "OFF":
		return log.OFF
	default:
		return log.ERROR
	}
}
