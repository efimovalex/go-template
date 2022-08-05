// Package mongodb provides functions to interact with mongo database.
package mongodb

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	maxIdleTimeMS = 120000
	retrywrites   = true
)

// Client represents a database client that contains functions used to interact with the database.
type Client struct {
	DB *mongo.Database
	*mongo.Client
	logger zerolog.Logger

	Users *mongo.Collection
}

// New - Creates a new Client from a sql.DB
func New(address, port, username, password, database string, ssl bool) (*Client, error) {
	var err error
	c := new(Client)
	c.logger = log.With().Str("component", "mongo").Logger()
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?retrywrites=%t&maxIdleTimeMS=%d&ssl=%t", username, password, address, port, retrywrites, maxIdleTimeMS, ssl)

	c.Client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri).SetDirect(true))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to db")
	}

	err = c.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping db")
	}

	c.DB = c.Client.Database(database)
	c.Users = c.DB.Collection(userCollection)

	c.logger.Info().Msgf("Connected to %s:%s", address, port)

	return c, nil
}

func (c *Client) Ping() error {
	return c.Client.Ping(context.Background(), nil)
}
