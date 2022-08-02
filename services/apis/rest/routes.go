package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ziflex/lecho/v3"
)

func (rest *R) SetupRouter() {
	r := echo.New()
	logger := lecho.From(rest.logger)
	r.Logger = logger

	// Add middlewarers
	r.Use(lecho.Middleware(lecho.Config{
		Logger: logger,
	}))
	r.Use(middleware.CORS())

	r.GET("/", rest.GetRoot)

	rest.Router = r
}
