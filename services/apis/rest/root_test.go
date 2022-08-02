package rest

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestREST_GetRoot(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		queryParams        map[string]string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "Success default name",
			body:               "",
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Hello, world!"}`,
		},
		{
			name:               "Success other name",
			body:               "",
			queryParams:        map[string]string{"name": "Alex"},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Hello, world!"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTestREST(t)

			e := echo.New()
			q := make(url.Values)
			if tt.queryParams != nil {
				for k, v := range tt.queryParams {
					q.Set(k, v)
				}
			}
			req, err := http.NewRequest("GET", "/?"+q.Encode(), strings.NewReader(tt.body))
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, w)

			err = r.GetRoot(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			checkResponseWithTestDataFile(t, w.Body.Bytes(), []string{})
		})
	}

	t.Run("Success without pretty response", func(t *testing.T) {
		r := NewTestREST(t)
		r.prettyResponse = false
		e := echo.New()

		req, err := http.NewRequest("GET", "/", strings.NewReader(""))
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, w)

		err = r.GetRoot(c)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)
		checkResponseWithTestDataFile(t, w.Body.Bytes(), []string{})
	})
}
