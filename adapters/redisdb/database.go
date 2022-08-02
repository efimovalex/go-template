package redisdb

import (

	// sql driver

	"context"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

// Client - database client
type Client struct {
	DB     *redis.Client
	logger zerolog.Logger
}

// New - Creates a new Client
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

func (c *Client) Ping() error {
	_, err := c.DB.Ping(context.Background()).Result()
	return err
}

func NewTestDB(t *testing.T) *Client {
	if t == nil {
		return nil
	}
	db, err := New("localhost", "6379", "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81", 15)
	assert.NoError(t, err)
	return db
}
