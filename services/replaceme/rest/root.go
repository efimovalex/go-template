package rest

import (
	"net/http"
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
// @Param Authorization header string true "Example: Bearer token"
// @Param data body string true "request JSON params"
// @Success 200 {object} string "No content"
// @Failure 400 {object} Message "Invalid request JSON"
// @Failure 422 {object} Message "Params validation error"
// @Failure 500 {object} Message "Internal server error"
// @Router / [get]
func (rest *R) GetRoot(resp http.ResponseWriter, r *http.Request) {
	rest.JSON(resp, http.StatusOK, Message{"Hello, world!"})
}
