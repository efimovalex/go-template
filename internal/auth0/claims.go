package auth

import (
	"context"
	"fmt"
	"strings"
)

const (
	claimsKeyEmail = "parsed_email"
)

// CustomClaims contains custom data from a JWT token.
type CustomClaims map[string]interface{}

// Validate is needed to satisfy Auth0 validator.CustomClaims interface.
func (cc CustomClaims) Validate(ctx context.Context) error {
	var errs []string

	userID, ok := cc["sub"]
	if !ok {
		errs = append(errs, "no user ID found in claims")
	}
	if ok {
		_, ok = userID.(string)
		if !ok {
			errs = append(errs, fmt.Sprintf(
				"user ID '%v' from claims has unexpected type: expected 'string', got '%T'",
				userID, userID))
		}
	}

	email, err := cc.parseEmail()
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to validate token claims: %s", strings.Join(errs, ", "))
	}

	cc[claimsKeyEmail] = email

	return nil
}

// GetUserID returns the user ID from the claims.
func (cc CustomClaims) GetUserID() string {
	return cc["sub"].(string)
}

// GetUserEmail returns the user email from the claims.
func (cc CustomClaims) GetUserEmail() string {
	return cc[claimsKeyEmail].(string)
}

func (cc CustomClaims) parseEmail() (string, error) {
	for k, v := range cc {
		if strings.HasSuffix(k, "email") {
			email, ok := v.(string)
			if !ok {
				return "", fmt.Errorf(
					"email '%v' from claims has unexpected type: expected 'string', got '%T'",
					v, v)
			}
			return email, nil
		}
	}
	return "", nil
}
