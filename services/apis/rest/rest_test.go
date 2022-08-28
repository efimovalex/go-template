package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/efimovalex/replaceme/adapters/postgres"
	auth "github.com/efimovalex/replaceme/internal/auth0"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// NewTestREST - creates new REST instance for testing
func NewTestREST(t *testing.T) *R {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	assert.NoError(t, err)
	mock.ExpectPing().WillReturnError(errors.New("ping error"))
	sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

	r := R{
		logger:         zerolog.Nop(),
		DB:             &postgres.Client{DB: sqlxMock},
		prettyResponse: true,
	}

	r.SetupRouter()

	r.AuthMiddleware, err = r.AuthMiddlewareSetup(auth.New("https://some-domain/", []string{"some-audience"}))
	assert.NoError(t, err)

	return &r
}

func TestREST_New(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	assert.NoError(t, err)
	mock.ExpectPing().WillReturnError(errors.New("ping error"))
	sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

	t.Run("test success", func(t *testing.T) {
		claims := auth.New("http://some-domain", []string{""})
		h, err := New(&postgres.Client{DB: sqlxMock}, nil, nil, claims, true, "9000")
		assert.NoError(t, err)
		assert.NotNil(t, h)
	})

	t.Run("test auth init error", func(t *testing.T) {
		claims := auth.New("\\", []string{""})
		h, err := New(&postgres.Client{DB: sqlxMock}, nil, nil, claims, true, "9000")
		assert.Error(t, err)
		assert.Nil(t, h)
	})
}

func TestHealth_StopGrecefully(t *testing.T) {
	t.Parallel()
	r := NewTestREST(t)
	r.SetupRouter()
	testServer := httptest.NewServer(r.Router)
	defer testServer.Close()
	r.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	time.Sleep(time.Millisecond * 5)
	<-ctx.Done()
	r.Stop(ctx)

	_, err = http.Get(testServer.URL + "/")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}

func TestHealth_StopDeadline(t *testing.T) {
	t.Parallel()
	r := NewTestREST(t)
	r.SetupRouter()
	testServer := httptest.NewServer(r.Router)
	defer testServer.Close()
	r.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 5)
	<-ctx.Done()
	r.Stop(ctx)

	_, err = http.Get(testServer.URL + "/")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}

func TestR_Start(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Log(testServer.URL)
	r := NewTestREST(t)
	r.srv = testServer.Config
	r.srv.Addr = "not:a:real:address"

	err := r.Start(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "listen tcp: address not:a:real:address: too many colons in address", err.Error())
}
