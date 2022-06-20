package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	auth "github.com/iconimpact/replaceme/internal/auth0"
	"github.com/iconimpact/replaceme/internal/sqldb"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func NewTestREST(t *testing.T) REST {
	r := R{
		logger: zap.NewNop().Sugar(),
		DB:     sqldb.NewTestDB(t),
	}

	var err error
	r.AuthMiddleware, err = r.AuthMiddlewareSetup(auth.New("https://some-domain/", []string{"some-audience"}, zap.NewNop().Sugar()))
	assert.NoError(t, err)

	return &r
}

func TestREST_New(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		claims := auth.New("http://some-domain", []string{""}, zap.NewNop().Sugar())
		h, err := New(sqldb.NewTestDB(nil), nil, nil, claims, "9000", zap.NewNop().Sugar())
		assert.NoError(t, err)
		assert.NotNil(t, h)
	})

	t.Run("test auth init error", func(t *testing.T) {
		claims := auth.New("\\", []string{""}, zap.NewNop().Sugar())
		h, err := New(sqldb.NewTestDB(nil), nil, nil, claims, "9000", zap.NewNop().Sugar())
		assert.Error(t, err)
		assert.Nil(t, h)
	})
}

func TestREST_Stop(t *testing.T) {
	h := NewTestREST(t)

	r := h.(*R)
	testServer := httptest.NewServer(http.HandlerFunc(r.GetRoot))
	defer testServer.Close()
	r.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)
	h.Stop()

	_, err = http.Get(testServer.URL + "/")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}
