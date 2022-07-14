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
		qname              string
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
			qname:              "Alex",
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

			if tt.qname != "" {
				setQueryParams(req, map[string][]string{"name": {tt.qname}})
			}

			setURLParams(req, map[string]string{"name": "replaceme"})

			r.GetRoot(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			checkResponseWithTestDataFile(t, w.Body.Bytes(), []string{})
		})
	}
}
