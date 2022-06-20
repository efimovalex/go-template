package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestREST_GetRoot(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "TestREST_GetRoot-Success",
			body:               "",
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Hello, world!"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTestREST(t).(*R)

			req, err := http.NewRequest("GET", "/", strings.NewReader(tt.body))
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			r.GetRoot(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}
