package sqldb

import (

	// sql driver
	"fmt"

	_ "github.com/lib/pq"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Client - database client
type Client struct {
	*sqlx.DB
	psql   func() sq.StatementBuilderType
	logger *zap.SugaredLogger
}

// New - Creates a new Client from a sql.DB
func New(host, port, user, password, database, ssl string, logger *zap.SugaredLogger) (*Client, error) {
	var err error
	c := new(Client)
	c.logger = logger

	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", host, port, user, password, ssl)

	if database != "" {
		conn += fmt.Sprintf(" dbname=%s", database)
	}

	c.DB, err = sqlx.Connect("postgres", conn)
	if err != nil {
		c.logger.Errorw("Unable to connect to database", "error", err)
		return nil, err
	}

	err = c.Ping()
	if err != nil {
		c.logger.Errorw("Unable to ping database", "error", err)
		return nil, err
	}
	c.psql = initPSQL
	c.logger.Infof("Connected to %s:%s", host, port)
	return c, nil
}

func (c *Client) Ping() error {
	return c.DB.Ping()
}

func initPSQL() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
