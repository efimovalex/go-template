package main

import (
	"io"
	"os"

	"github.com/efimovalex/replaceme/config"
	server "github.com/efimovalex/replaceme/services/replaceme"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	exitOK   = 0
	exitFail = 1
)

// main - main entry point that loads configuration and starts the services
func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		log.Error().Err(err).Msg("failed to run")
		os.Exit(exitFail)
	}
	os.Exit(exitOK)
}

func run(args []string, stdout io.Writer) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	cfg, err := config.Load()
	if err != nil {
		log.Error().Err(err).Msg("config load error")

		return err
	}
	if cfg.Logger.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Msgf("config loaded: %+v", cfg)

	// start services
	server, err := server.New(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed init replaceme service")

		return err
	}

	server.Start()

	return nil
}
