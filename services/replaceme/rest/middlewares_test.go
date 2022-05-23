package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestREST_LogRequestMiddleware(t *testing.T) {
	r := initTestREST(t)
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
		{
			name: "Test LogRequestMiddleware",
			args: args{
				next: http.HandlerFunc(r.GetRoot),
			},
			want: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := r.LogRequestMiddleware(tt.args.next)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			assert.Equal(t, `{"message":"Hello, world!"}`, w.Body.String())
		})
	}
}

func Test_addTimeContextMiddleware(t *testing.T) {
	r := initTestREST(t)
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
		{
			name: "Test addTimeContextMiddleware",
			args: args{
				next: http.HandlerFunc(r.GetRoot),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := addTimeContextMiddleware(tt.args.next)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			h.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			assert.Equal(t, `{"message":"Hello, world!"}`, w.Body.String())
		})
	}
}
