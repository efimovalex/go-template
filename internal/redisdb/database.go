package redisdb

import (

	// sql driver

	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Client - database client
type Client struct {
	DB     *redis.Client
	logger *zap.SugaredLogger
}

// New - Creates a new Client
func New(host, port, password string, database int, logger *zap.SugaredLogger) (*Client, error) {
	var err error
	c := new(Client)
	c.logger = logger

	c.DB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       database, // use default DB
	})

	err = c.Ping()
	if err != nil {
		c.logger.Errorw("Unable to ping database", "error", err)
		return nil, err
	}

	c.logger.Infof("Connected to %s:%s", host, port)
	return c, nil
}

func (c *Client) Ping() error {
	_, err := c.DB.Ping(context.Background()).Result()
	return err
}
