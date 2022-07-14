package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/efimovalex/replaceme/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg, err := config.Load()
	assert.NoError(t, err)

	tests := []struct {
		name    string
		config  *config.Config
		wantErr error
	}{
		{
			name: "success",
			config: func() *config.Config {
				cfg2 := *cfg
				cfg2.Postgres.Name = "replaceme_test"

				return &cfg2
			}(),
		},
		{
			name: "ERROR db connection",
			config: func() *config.Config {
				cfg2 := *cfg
				cfg2.Postgres.User = ""
				cfg2.Postgres.Password = ""

				return &cfg2
			}(),
			wantErr: errors.New(`pq: password authentication failed for user "password="`),
		},
		{
			name: "ERROR auth issuer issues",
			config: func() *config.Config {
				cfg2 := *cfg
				cfg2.Auth.Domain = "\\"
				cfg2.Auth.Audience = ""

				return &cfg2
			}(),
			wantErr: errors.New(`failed to parse the issuer url: parse "https://\\/": invalid character "\\" in host name`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			server, err := New(tt.config)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, tt.wantErr)
				assert.NotNil(t, server.DB)
				assert.NotNil(t, server.REST)
				assert.Equal(t, server.cfg, tt.config)
			}
		})
	}
}

type SMock struct {
	testServer *httptest.Server
}

func (h *SMock) Check(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func (h *SMock) Start() {}
func (h *SMock) Stop() {
	h.testServer.Close()
}

func TestServer_Start(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "successful shutdown on SIGKILL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := new(SMock)
			h := new(SMock)

			r.testServer = httptest.NewServer(nil)
			h.testServer = httptest.NewServer(http.HandlerFunc(h.Check))

			s := &Server{
				sigChan:     make(chan os.Signal, 1),
				HealthCheck: h,
				REST:        r,
				cfg:         &config.Config{},
			}

			go s.Start()

			time.Sleep(150 * time.Millisecond)

			resp, err := http.Get(h.testServer.URL + "/healthcheck")
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, http.StatusOK)
			s.sigChan <- os.Kill

			_, err = http.Get(h.testServer.URL + "/healthcheck")
			assert.Error(t, err)
			assert.Equal(t, `Get "`+h.testServer.URL+`/healthcheck": dial tcp 127.0.0.1:`+strings.Split(h.testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
		})
	}
}
