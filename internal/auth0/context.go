package auth

import (
	"context"
	"fmt"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

type ctxKey string

var (
	ctxKeyUserID    = ctxKey("user-id")
	ctxKeyUserEmail = ctxKey("user-email")
)

// ClaimsValue returns the JWT claims from the specified context.
func ClaimsValue(ctx context.Context) (*CustomClaims, error) {
	contextClaims := ctx.Value(jwtmiddleware.ContextKey{})
	if contextClaims == nil {
		return nil, fmt.Errorf("no auth claims found in context")
	}

	validatedClaims, ok := contextClaims.(*validator.ValidatedClaims)
	if !ok {
		return nil, fmt.Errorf(
			"context has unexpected auth claims type: "+
				"expected *ValidatedClaims, got %T", contextClaims)
	}

	if validatedClaims.CustomClaims == nil {
		return nil, fmt.Errorf(
			"validated auth claims from context have no custom claims")
	}

	customClaims, ok := validatedClaims.CustomClaims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf(
			"custom auth claims from context have unexpected type: "+
				"expected *CustomClaims, got %T", validatedClaims.CustomClaims)
	}

	return customClaims, nil
}

// WithUserID sets the user ID in the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

// UserIDValue retrieves the user ID from the context.
func UserIDValue(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(ctxKeyUserID).(string)
	if !ok {
		return "", fmt.Errorf("no user ID found in context")
	}
	return userID, nil
}

// WithUserEmail sets the user email in the context.
func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, ctxKeyUserEmail, ctxKeyUserEmail)
}

// UserEmailValue retrieves the user email from the context.
func UserEmailValue(ctx context.Context) (string, error) {
	userEmail, ok := ctx.Value(ctxKeyUserEmail).(string)
	if !ok {
		return "", fmt.Errorf("no user email found in context")
	}
	return userEmail, nil
}
