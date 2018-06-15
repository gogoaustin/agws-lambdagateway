package app

import (
	"net/http"

	"git.gogoair.com/bagws/lambdagateway/app/payment"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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

	// Register payment routes
	payment.Register(e)

	return e
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Everything is awesome!")
}
