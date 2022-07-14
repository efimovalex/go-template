package config

import (
	// to embed default config
	"context"
	_ "embed"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
)

// Config used globally
type Config struct {
	Logger      Logger      `env:",prefix=LOG_"`
	REST        REST        `env:",prefix=REST_"`
	Static      Static      `env:",prefix=STATIC_"`
	HealthCheck HealthCheck `env:",prefix=HC_"`

	Postgres Postgres `env:",prefix=POSTGRES_"`
	Mongo    Mongo    `env:",prefix=MONGO_"`
	Redis    Redis    `env:",prefix=REDIS_"`

	Auth Auth `env:",prefix=AUTH_"`

	Swagger Swagger `env:",prefix=SWAGGER_"`

	Logging Logging
}

type Swagger struct {
	Enable bool   `env:"ENABLE,default=true"`
	Port   string `env:"PORT,default=8085"`
}

type Logger struct {
	Level  string `env:"LEVEL,default=info"`
	Pretty bool   `env:"PRETTY,default=false"`
}

// REST configuration
type REST struct {
	Port   string `env:"PORT,default=8080"`
	Pretty bool   `env:"PRETTY,default=false"`
}

// HealthCheck configuration
type HealthCheck struct {
	Port string `env:"PORT,default=8081"`
}

// Static configuration
type Static struct {
	Port string `env:"PORT,default=8083"`
}

// Logging configuration
type Logging struct {
	Level       string `env:"LEVEL,default=info"`
	Development bool   `env:"DEV,default=false"`
}

// Postgres configuration
type Postgres struct {
	Host     string `env:"HOST,default=localhost"`
	Name     string `env:"NAME,default=replaceme"`
	User     string `env:"USER,default=replaceme"`
	Password string `env:"PASSWORD,default=replaceme"`
	Port     string `env:"PORT,default=5433"`
	SSLMode  string `env:"SSL_MODE,default=disable"`
}

// Mongo configuration
type Mongo struct {
	Host     string `env:"HOST,default=localhost"`
	Name     string `env:"NAME,default=mongo_db"`
	User     string `env:"USER,default=root"`
	Password string `env:"PASSWORD,default=root"`
	Port     string `env:"PORT,default=27017"`
	SSLMode  bool   `env:"SSL_MODE,default=false"`
}

// Redis configuration
type Redis struct {
	Host     string `env:"HOST,default=localhost"`
	Database int    `env:"NAME,default=0"`
	User     string `env:"USER,default=root"`
	Password string `env:"PASSWORD,default=eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81"` // this is from docker-compose.yml not a real password
	Port     string `env:"PORT,default=6379"`
}

type Auth struct {
	Domain   string `env:"DOMAIN,default=replaceme.eu.auth0.com"`
	Audience string `env:"AUDIENCE,default=https://replaceme.com"`
}

// Load reads info from TOML file at relative path
func Load() (*Config, error) {
	var c Config
	ctx := context.Background()
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to decode default config")
	}

	return &c, nil
}
