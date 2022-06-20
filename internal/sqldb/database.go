package sqldb

import (

	// sql driver
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

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

func NewTestDB(t *testing.T) *Client {
	if t == nil {
		return nil
	}
	db, err := New("localhost", "5432", "replaceme", "replaceme", "replaceme_test", "disable", zap.NewNop().Sugar())
	assert.NoError(t, err)
	return db
}

// USE ONLY FOR TESTING
func (db *Client) resetTable(table string) error {
	_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	if err != nil {
		return err
	}
	return nil
}
