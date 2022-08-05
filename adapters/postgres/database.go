// Package postgres provides functions to interact with postgres database.
package postgres

import (

	// sql driver
	"fmt"
	"reflect"
	"regexp"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/jmoiron/sqlx"
)

const driverName = "postgres"

// Client represents a database client that contains functions used to interact with the database.
type Client struct {
	*sqlx.DB
	logger zerolog.Logger
}

// New creates a new postgres.Client
func New(host, port, user, password, database, ssl string) (*Client, error) {
	var err error
	c := new(Client)
	c.logger = log.With().Str("component", driverName).Logger()

	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", host, port, user, password, ssl)

	if database != "" {
		conn += fmt.Sprintf(" dbname=%s", database)
	}

	c.DB, err = sqlx.Connect(driverName, conn)
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

// Ping checks if the database is reachable
func (c *Client) Ping() error {
	return c.DB.Ping()
}

// logQuery logs the query and its parameter values
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
			if driverName == "mysql" {
				if rv.Bool() {
					a = append(a, 1)
				} else {
					a = append(a, 0)
				}
			} else {
				a = append(a, fmt.Sprintf(`%t`, rv.Bool()))
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
