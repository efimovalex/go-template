package auth

import (
	"context"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/stretchr/testify/require"
)

type fakeCustomClaims struct{}

func (fcc *fakeCustomClaims) Validate(ctx context.Context) error {
	return nil
}

func TestClaimsValue(t *testing.T) {

	testCases := []struct {
		name      string
		ctx       context.Context
		wantError string
	}{
		{
			"invalid_claims_type",
			context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, ""),
			"context has unexpected auth claims type: expected *ValidatedClaims, got string",
		},
		{
			"no_custom_claims",
			context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, &validator.ValidatedClaims{}),
			"validated auth claims from context have no custom claims",
		},
		{
			"no_custom_claims",
			context.WithValue(
				context.Background(),
				jwtmiddleware.ContextKey{},
				&validator.ValidatedClaims{CustomClaims: &fakeCustomClaims{}},
			),
			"custom auth claims from context have unexpected type: " +
				"expected *CustomClaims, got *auth.fakeCustomClaims",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ClaimsValue(tc.ctx)
			require.EqualError(t, err, tc.wantError)
		})
	}
}

func TestContextUserID(t *testing.T) {
	// missing user ID
	ctx := context.Background()
	_, err := UserIDValue(ctx)
	require.EqualError(t, err, "no user ID found in context")

	// happy path
	userID := "auth0|1a2b3c4d5e6f7g8h9i0a1b2a"
	ctx = WithUserID(ctx, userID)
	actualUserID, err := UserIDValue(ctx)
	require.NoError(t, err)
	require.Equal(t, userID, actualUserID)

}
