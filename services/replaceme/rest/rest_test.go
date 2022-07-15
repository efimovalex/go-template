package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/efimovalex/replaceme/internal/postgres"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func NewTestREST(t *testing.T) *R {
	r := R{
		logger:         zerolog.Nop(),
		DB:             postgres.NewTestDB(t),
		prettyResponse: true,
	}

	r.SetupRouter()

	var err error
	r.AuthMiddleware, err = r.AuthMiddlewareSetup(auth.New("https://some-domain/", []string{"some-audience"}))
	assert.NoError(t, err)

	return &r
}

func TestREST_New(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		claims := auth.New("http://some-domain", []string{""})
		h, err := New(postgres.NewTestDB(nil), nil, nil, claims, true, "9000")
		assert.NoError(t, err)
		assert.NotNil(t, h)
	})

	t.Run("test auth init error", func(t *testing.T) {
		claims := auth.New("\\", []string{""})
		h, err := New(postgres.NewTestDB(nil), nil, nil, claims, true, "9000")
		assert.Error(t, err)
		assert.Nil(t, h)
	})
}

func TestREST_Stop(t *testing.T) {
	r := NewTestREST(t)
	r.SetupRouter()
	testServer := httptest.NewServer(r.Router)
	defer testServer.Close()
	r.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)
	r.Stop()

	_, err = http.Get(testServer.URL + "/")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}
