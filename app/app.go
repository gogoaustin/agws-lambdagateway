package app

import (
	"net/http"

	"git.gogoair.com/bagws/lambdagateway/app/payment"
	"github.com/labstack/echo"
)

// NewApp creates and returns a new router
func NewApp() *echo.Echo {
	e := echo.New()
	e.GET("/healthcheck", healthHandler)

	// Register payment routes
	payment.Register(e)

	return e
}

func healthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Everything is awesome!")
}
