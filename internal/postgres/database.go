package postgres

import (

	// sql driver
	"fmt"
	"reflect"
	"regexp"
	"testing"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/jmoiron/sqlx"
)

// Client - database client
type Client struct {
	*sqlx.DB
	logger zerolog.Logger
}

// New - Creates a new Client from a sql.DB
func New(host, port, user, password, database, ssl string) (*Client, error) {
	var err error
	c := new(Client)
	c.logger = log.With().Str("component", "postgres").Logger()

	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", host, port, user, password, ssl)

	if database != "" {
		conn += fmt.Sprintf(" dbname=%s", database)
	}

	c.DB, err = sqlx.Connect("postgres", conn)
	if err != nil {
		c.logger.Error().Err(err).Msg("Unable to connect to database")
		return nil, err
	}

	err = c.Ping()
	if err != nil {
		c.logger.Error().Err(err).Msg("Unable to ping database")
		return nil, err
	}
	c.logger.Info().Msgf("Connected to %s:%s", host, port)
	return c, nil
}

func (c *Client) Ping() error {
	return c.DB.Ping()
}

func NewTestDB(t *testing.T) *Client {
	if t == nil {
		return nil
	}
	db, err := New("localhost", "5433", "replaceme", "replaceme", "replaceme_test", "disable")
	assert.NoError(t, err)
	return db
}

// ResetTable - USE ONLY FOR TESTING to reset tables
func (db *Client) ResetTable(t *testing.T, table string) {
	_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	assert.NoError(t, err)
}

func (db *Client) logQuery(query string,
	args ...interface{}) {
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")
	q := regexp.MustCompile(`\$\d`).ReplaceAllString(query, "%v")
	a := []interface{}{}
	for _, i := range args {
		rv := reflect.ValueOf(i)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Bool {
			if rv.Bool() {
				a = append(a, 1)
			} else {
				a = append(a, 0)
			}
			continue
		}

		if rv.Kind() == reflect.String {
			a = append(a, fmt.Sprintf(`'%s'`, rv))
			continue
		}

		a = append(a, rv)

	}

	db.logger.Debug().Msgf(q, a...)
}
