package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	auth "github.com/efimovalex/replaceme/internal/auth0"

	"github.com/stretchr/testify/assert"
)

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
			r := NewTestREST(t)
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
			w := httptest.NewRecorder()

			got.CheckJWT(handlerFunc).ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code, "body:\n%s", w.Body.String())

			checkResponseWithTestDataFile(t, w.Body.Bytes(), []string{})
		})
	}
}
