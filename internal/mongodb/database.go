package mongodb

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	maxIdleTimeMS = 120000
	retrywrites   = true
)

// Client - database client
type Client struct {
	DB *mongo.Database
	*mongo.Client
	logger *zap.SugaredLogger
}

// New - Creates a new Client from a sql.DB
func New(address, port, username, password, database string, ssl bool, logger *zap.SugaredLogger) (*Client, error) {
	var err error
	c := new(Client)
	c.logger = logger
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

	c.logger.Infof("Connected to %s:%s", address, port)
	return c, nil
}

func (c *Client) Ping() error {
	return c.Client.Ping(context.Background(), nil)
}
