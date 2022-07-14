package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/rs/zerolog"

	"github.com/stretchr/testify/assert"
)

func TestREST_LogRequestMiddleware(t *testing.T) {
	r := NewTestREST(t).(*R)
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
	r := NewTestREST(t).(*R)
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

func Test_CORSMiddleware(t *testing.T) {
	r := NewTestREST(t).(*R)
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		verb string
		args args
		want http.Handler
	}{
		{
			name: "Test CORSMiddleware",
			verb: "GET",
			args: args{
				next: http.HandlerFunc(r.GetRoot),
			},
		},
		{
			name: "Test CORSMiddleware OPTIONS",
			verb: "OPTIONS",
			args: args{
				next: http.HandlerFunc(r.GetRoot),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := CORSMiddleware(tt.args.next)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.verb, "/", nil)
			h.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			assert.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), "*")
			assert.Equal(t, w.Header().Get("Access-Control-Allow-Methods"), "POST, GET, OPTIONS, PUT, PATCH, DELETE")
			assert.Equal(t, w.Header().Get("Access-Control-Allow-Headers"), "Accept, Authorization, Accept-Language, Accept-Encoding, "+
				"Content-Type, Content-Length, X-CSRF-Token, Session-UUID, Device-UUID")
			assert.Equal(t, w.Header().Get("Access-Control-Allow-Credentials"), "true")
		})
	}
}

func TestUserCtx(t *testing.T) {
	r := NewTestREST(t).(*R)
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "http://some-test-url", nil)

	t.Run("invalid auth claims", func(t *testing.T) {
		// invalid auth claims
		respRecorder := httptest.NewRecorder()
		r.UserCtx(handlerFunc).ServeHTTP(respRecorder, req)
		respBody := respRecorder.Body.String()
		assert.Equal(
			t, http.StatusUnauthorized, respRecorder.Code, "body:\n%s", respBody)
		assert.Equal(t, `{"message":"no auth claims found in context"}`, respBody)

	})

	t.Run("valid auth claims", func(t *testing.T) {
		// happy path: set claims in request context
		claimsWithWriteAll := &validator.ValidatedClaims{
			CustomClaims: &auth.CustomClaims{
				"sub":                       "auth0|1a2b3c4d5e6f7g8h9i0a1b2c",
				"https://some-domain/email": "some-email@some-domain.com",
			},
		}
		assert.NoError(t, claimsWithWriteAll.CustomClaims.Validate(context.Background()))

		ctx := context.WithValue(req.Context(), jwtmiddleware.ContextKey{}, claimsWithWriteAll)
		req = req.WithContext(ctx)
		respRecorder := httptest.NewRecorder()
		r.UserCtx(handlerFunc).ServeHTTP(respRecorder, req)
		respBody := respRecorder.Body.String()
		assert.Equal(t, http.StatusOK, respRecorder.Code, "body:\n%s", respBody)
	})
}

func TestR_AuthMiddlewareSetup(t *testing.T) {
	tests := []struct {
		name    string
		want    *jwtmiddleware.JWTMiddleware
		wantErr bool
	}{
		{
			name: "success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &R{
				logger: zerolog.Nop(),
			}
			got, err := r.AuthMiddlewareSetup(auth.New("http://some-domain", []string{"some-audience"}))
			if (err != nil) != tt.wantErr {
				t.Errorf("R.AuthMiddlewareSetup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)

			handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// happy path: set claims in request context
			claimsWithWriteAll := &validator.ValidatedClaims{
				CustomClaims: &auth.CustomClaims{
					"sub":                       "auth0|1a2b3c4d5e6f7g8h9i0a1b2c",
					"https://some-domain/email": "some-email@some-domain.com",
				},
			}
			assert.NoError(t, claimsWithWriteAll.CustomClaims.Validate(context.Background()))

			req := httptest.NewRequest("GET", "http://some-test-url", nil)
			ctx := context.WithValue(req.Context(), jwtmiddleware.ContextKey{}, claimsWithWriteAll)
			req = req.WithContext(ctx)
			respRecorder := httptest.NewRecorder()

			got.CheckJWT(handlerFunc).ServeHTTP(respRecorder, req)

			respBody := respRecorder.Body.String()

			assert.Equal(t, http.StatusUnauthorized, respRecorder.Code, "body:\n%s", respBody)
		})
	}
}
