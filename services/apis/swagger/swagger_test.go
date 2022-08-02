package swagger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/efimovalex/replaceme/docs/swagger"
	"github.com/stretchr/testify/assert"
)

func TestS_StopGraceful(t *testing.T) {

	h := New("8085", "3000")

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	h.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/healthcheck")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	<-ctx.Done()
	time.Sleep(time.Millisecond * 5)
	h.Stop(ctx)

	_, err = http.Get(testServer.URL + "/healthcheck")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/healthcheck": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}

func TestS_StopError(t *testing.T) {
	h := New("8085", "3000")

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	h.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/healthcheck")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 5)
	<-ctx.Done()
	h.Stop(ctx)

	_, err = http.Get(testServer.URL + "/healthcheck")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/healthcheck": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}

func TestS_Start(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Log(testServer.URL)
	h := New("not_a_port", "8080")

	err := h.Start(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "listen tcp: lookup tcp/not_a_port: nodename nor servname provided, or not known", err.Error())
}
