// Package redisdb provides functions to interact with redis database.
package redisdb

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client represents a database client that contains functions used to interact with the database.
type Client struct {
	DB     *redis.Client
	logger zerolog.Logger
}

// New creates a new redisdb.Client
func New(host, port, password string, database int) (*Client, error) {
	var err error
	c := new(Client)

	c.logger = log.With().Str("component", "redis").Logger()

	c.DB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       database, // use default DB
	})

	err = c.Ping()
	if err != nil {
		c.logger.Error().Err(err).Msg("Unable to ping database")
		return nil, err
	}

	c.logger.Info().Msgf("Connected to %s:%s", host, port)
	return c, nil
}

// Ping pings the database to check if it is alive.
func (c *Client) Ping() error {
	_, err := c.DB.Ping(context.Background()).Result()
	return err
}
