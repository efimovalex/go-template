package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (rest *R) SetupRouter() {
	r := echo.New()

	// Add middlewarers
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Add routes
	r.GET("/", rest.GetRoot)

	rest.Router = r
}
