// Package swagger starts a swagger UI server that displays the swagger documentation of the API service
package swagger

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	_ "github.com/efimovalex/replaceme/docs/swagger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
)

// S is a swagger service implementation
type S struct {
	logger zerolog.Logger
	srv    *http.Server
}

// New creates a new swagger service
func New(port string, apiPort string) *S {
	h := &S{
		srv: &http.Server{Addr: "0.0.0.0:" + port},

		logger: log.With().Str("component", "Swagger").Logger(),
	}

	uri, _ := url.Parse(fmt.Sprintf("http://localhost:%s/", apiPort))

	h.srv.Handler = httpSwagger.Handler(
		httpSwagger.URL("0.0.0.0:"+port+"/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.PersistAuthorization(true),
		httpSwagger.BeforeScript(`const UrlMutatorPlugin = (system) => ({
			rootInjects: {
			  setScheme: (scheme) => {
				const jsonSpec = system.getState().toJSON().spec.json;
				const schemes = Array.isArray(scheme) ? scheme : [scheme];
				const newJsonSpec = Object.assign({}, jsonSpec, { schemes });
		  
				return system.specActions.updateJsonSpec(newJsonSpec);
			  },
			  setHost: (host) => {
				const jsonSpec = system.getState().toJSON().spec.json;
				const newJsonSpec = Object.assign({}, jsonSpec, { host });
		  
				return system.specActions.updateJsonSpec(newJsonSpec);
			  },
			  setBasePath: (basePath) => {
				const jsonSpec = system.getState().toJSON().spec.json;
				const newJsonSpec = Object.assign({}, jsonSpec, { basePath });
		  
				return system.specActions.updateJsonSpec(newJsonSpec);
			  }
			}
		  });`),
		httpSwagger.Plugins([]string{"UrlMutatorPlugin"}),
		httpSwagger.UIConfig(map[string]string{
			"onComplete": fmt.Sprintf(`() => {
			window.ui.setScheme('%s');
			window.ui.setHost('%s');
			window.ui.setBasePath('%s');
		}`, uri.Scheme, uri.Host, uri.Path),
		}),
	)

	return h
}

// Start starts the swagger service
func (h *S) Start(ctx context.Context) error {
	h.logger.Info().Msgf("Starting swagger service http://%s", h.srv.Addr)
	h.logger.Info().Msgf("Documentation url:  http://%s/swagger/index.html", h.srv.Addr)
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", h.srv.Addr)
	if err != nil {
		return err
	}
	if err := h.srv.Serve(ln); err != http.ErrServerClosed {
		// Error starting or closing listener:
		h.logger.Fatal().Msgf("Swagger server error: %v", err)

		return err
	}

	return nil
}

// Stop stops the swagger service
func (h *S) Stop(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := h.srv.Shutdown(ctx); err != nil {
		if err == context.Canceled {
			h.logger.Info().Msg("Swagger server shutdown gracefully")
		} else {
			h.logger.Error().Msgf("Swagger server error: %v", err)
		}
	}
}
