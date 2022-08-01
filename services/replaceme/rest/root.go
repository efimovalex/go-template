package rest

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Message is a simple JSON response
type Message struct {
	Message string `json:"message"`
}

// @Summary [get] /
// @Description Returns root endpoint
// @Tags root
// @Accept  json
// @Produce json
// @Param data body string true "request JSON params"
// @Success 200 {object} string "No content"
// @Failure 400 {object} Message "Invalid request JSON"
// @Failure 422 {object} Message "Params validation error"
// @Failure 500 {object} Message "Internal server error"
// @Router / [get]
func (rest *R) GetRoot(c echo.Context) error {
	name := c.QueryParams().Get("name")
	if name == "" {
		name = "World"
	}

	return rest.JSON(c, http.StatusOK, Message{fmt.Sprintf("Hello, %s!", name)})
}
