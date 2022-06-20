package auth

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaims(t *testing.T) {
	claimsJSON := `{
    "https://icondevhvo.com/email": "some.user@some-non-existent-domain.com",
    "iss": "https://icondevhvo.eu.auth0.com/",
    "sub": "auth0|1a2b3c4d5e6f7g8h9i0a1b2c",
    "aud": [
      "https://icondevhvo.com",
      "https://icondevhvo.eu.auth0.com/userinfo"
    ],
    "iat": 1651239912,
    "exp": 1651247112,
    "azp": "OyVwQbOiVUsykbiWL1kmnv5IYCTdPBM8",
    "scope": "openid email"
  }`

	var claims CustomClaims
	assert.NoError(t, json.Unmarshal([]byte(claimsJSON), &claims))
	assert.NoError(t, claims.Validate(context.Background()))
	assert.Equal(t, "auth0|1a2b3c4d5e6f7g8h9i0a1b2c", claims.GetUserID())
	assert.Equal(t, "some.user@some-non-existent-domain.com", claims.GetUserEmail())

	// no user ID
	err := CustomClaims{}.Validate(context.Background())
	assert.Error(t, err)
	assert.Equal(
		t,
		"failed to validate token claims: no user ID found in claims",
		err.Error())

	// user ID and email are present, but they are not strings
	err = CustomClaims{
		"sub":                          true,
		"https://icondevhvo.com/email": false,
	}.Validate(context.Background())
	assert.Error(t, err)
	assert.Equal(
		t,
		"failed to validate token claims: "+
			"user ID 'true' from claims has unexpected type: expected 'string', got 'bool', "+
			"email 'false' from claims has unexpected type: expected 'string', got 'bool'",
		err.Error())
}
