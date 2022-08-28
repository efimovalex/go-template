package config

import (
	// to embed default config
	"context"
	_ "embed"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
)

// Config used at the start of the application
type Config struct {
	Logger      Logger      `env:",prefix=LOG_"`
	REST        REST        `env:",prefix=REST_"`
	HealthCheck HealthCheck `env:",prefix=HC_"`

	Postgres Postgres `env:",prefix=POSTGRES_"`
	Mongo    Mongo    `env:",prefix=MONGO_"`
	Redis    Redis    `env:",prefix=REDIS_"`

	Auth Auth `env:",prefix=AUTH_"`

	Swagger Swagger `env:",prefix=SWAGGER_"`
}

// Swagger represents the swagger service configuration
type Swagger struct {
	Enable bool   `env:"ENABLE,default=true"`
	Port   string `env:"PORT,default=8085"`
}

// Logger represents the logger configuration
type Logger struct {
	Level  string `env:"LEVEL,default=info"`
	Pretty bool   `env:"PRETTY,default=false"`
}

// REST represents the REST service configuration
type REST struct {
	Port   string `env:"PORT,default=8080"`
	Pretty bool   `env:"PRETTY,default=false"`
}

// HealthCheck represents the healthcheck service configuration
type HealthCheck struct {
	Port string `env:"PORT,default=8081"`
}

// Postgres represents the postgres service configuration
type Postgres struct {
	Host     string `env:"HOST,default=localhost"`
	Name     string `env:"NAME,default=replaceme"`
	User     string `env:"USER,default=replaceme"`
	Password string `env:"PASSWORD,default=replaceme"`
	Port     string `env:"PORT,default=5433"`
	SSLMode  string `env:"SSL_MODE,default=disable"`
}

// Mongo represents the mongo service configuration
type Mongo struct {
	Host     string `env:"HOST,default=localhost"`
	Name     string `env:"NAME,default=mongo_db"`
	User     string `env:"USER,default=root"`
	Password string `env:"PASSWORD,default=root"`
	Port     string `env:"PORT,default=27017"`
	SSLMode  bool   `env:"SSL_MODE,default=false"`
}

// Redis represents the redis service configuration
type Redis struct {
	Host     string `env:"HOST,default=localhost"`
	Database int    `env:"NAME,default=0"`
	User     string `env:"USER,default=root"`
	Password string `env:"PASSWORD,default=eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81"` // this is from docker-compose.yml not a real password
	Port     string `env:"PORT,default=6379"`
}

// Auth represents the auth middleware configuration
type Auth struct {
	Domain   string `env:"DOMAIN,default=replaceme.eu.auth0.com"`
	Audience string `env:"AUDIENCE,default=https://replaceme.com"`
}

// Load reads info from ENV and returns a Config struct
func Load() (*Config, error) {
	var c Config
	ctx := context.Background()
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, errors.Wrapf(err, "failed to decode default config")
	}

	return &c, nil
}
