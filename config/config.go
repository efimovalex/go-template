package config

import (
	// to embed default config
	_ "embed"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Config used globally
type Config struct {
	REST        REST
	Static      Static
	HealthCheck HealthCheck

	Database Database
	Mongo    Mongo
	Redis    Redis

	Logging Logging
}

// REST configuration
type REST struct {
	Port    string
	Timeout int
}

// Static configuration
type Static struct {
	Port    string
	Timeout int
}

// HealthCheck configuration
type HealthCheck struct {
	Port    string
	Timeout int
}

// Logging configuration
type Logging struct {
	Level       string
	Development bool
}

// Database configuration
type Database struct {
	Host     string
	Name     string
	User     string
	Password string
	Port     string
	SSLMode  string
}

// Mongo configuration
type Mongo struct {
	Host     string
	Name     string
	User     string
	Password string
	Port     string
	SSL      bool
}

// Redis configuration
type Redis struct {
	Host     string
	Port     string
	Password string
	Database int
}

//go:embed config_dev.toml
var configDev string

// LoadDefaultConfig loads the default config
func LoadDefaultConfig() (*Config, error) {
	var c Config
	_, err := toml.Decode(configDev, &c)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode default config")
	}

	return &c, nil
}

// Load reads info from TOML file at relative path
func Load(filename string) (*Config, error) {
	configFile := filename
	_, err := os.Stat(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "config file is missing: %s", configFile)
	}

	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return nil, errors.Wrapf(err, "parsing config file: %s", configFile)
	}

	/*
		// just do the secret value injection on production
		if config.Server.Env == "prod" {
			config = addSecrets(config)
		}*/

	return &config, nil
}
