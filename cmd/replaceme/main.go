package main

import (
	"fmt"
	"io"
	"os"

	"github.com/iconimpact/replaceme/config"
	server "github.com/iconimpact/replaceme/services/replaceme"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("config load error %s", err.Error())
		return err
	}

	log, err := loadLogger(cfg)
	if err != nil {
		fmt.Printf("failed to init logger: %s", err.Error())

		return err
	}
	defer func() { _ = log.Sync() }()

	log.Infow("config loaded", "config", cfg)

	// start services
	server, err := server.New(cfg, log)
	if err != nil {
		log.Errorf("failed init replaceme service: %s", err.Error())

		return err
	}

	server.Start()

	return nil
}

func loadConfig() (*config.Config, error) {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		return nil, fmt.Errorf("config file not specified - CONFIG_FILE env var must be set")
	}

	return config.Load(configFile)
}

func loadLogger(cfg *config.Config) (*zap.SugaredLogger, error) {
	var err error
	logConfig := zap.NewProductionConfig()
	logConfig.Level, err = zap.ParseAtomicLevel(cfg.Logging.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %s", err.Error())
	}
	logConfig.Development = cfg.Logging.Development
	logConfig.EncoderConfig.MessageKey = "message"
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := logConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %s", err.Error())
	}

	return logger.Sugar(), nil
}
