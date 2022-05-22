package rest

import (
	"net/http"

	"github.com/iconimpact/go-core/respond"
)

// @Summary [get] /
// @Description Returns root endpoint
// @Tags root
// @Accept  json
// @Produce json
// @Param Authorization header string true "Example: Bearer token"
// @Param data body struct{} true "request JSON params"
// @Success 200 {object} string "No content"
// @Failure 400 {object} errutil.Response "Invalid request JSON"
// @Failure 422 {object} errutil.Response "Params validation error"
// @Failure 500 {object} errutil.Response "Internal server error"
// @Router /api/fvt/digitaltwin [post]
func (r *REST) GetRoot(resp http.ResponseWriter, req *http.Request) {
	respond.JSON(resp, r.logger.Desugar(), http.StatusOK, map[string]string{"message": "Hello, world!"})
}
