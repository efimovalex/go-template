package main

import (
	"fmt"
	"io"
	"os"

	"github.com/iconimpact/replaceme/config"
	"github.com/iconimpact/replaceme/services/static"
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
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
	os.Exit(exitOK)
}

func run(args []string, stdout io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		log.Error().Err(err).Msg("config load error")

		return err
	}

	if cfg.Logger.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	static, err := static.New(cfg)
	if err != nil {
		log.Error().Err(err).Msgf("failed init static service")

		return err
	}

	static.Start()

	return nil
}
