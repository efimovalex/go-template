package healthcheck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/efimovalex/replaceme/internal/mongodb"
	"github.com/efimovalex/replaceme/internal/redisdb"
	"github.com/efimovalex/replaceme/internal/sqldb"
	"github.com/go-redis/redismock/v8"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestHealth_Check(t *testing.T) {
	db := sqldb.NewTestDB(t)

	mdb := mongodb.NewTestDB(t)

	rdb, err := redisdb.New("localhost", "6379", "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81", 15)
	assert.NoError(t, err)

	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	mock.ExpectPing().WillReturnError(errors.New("ping error"))
	sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

	defer mockDB.Close()

	redisClientMock, redismock := redismock.NewClientMock()
	redismock.ExpectPing().SetErr(errors.New("redis error"))
	defer redisClientMock.Close()

	tests := []struct {
		name           string
		DB             *sqldb.Client
		Mongo          *mongodb.Client
		Redis          *redisdb.Client
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "Test health check success",
			DB:             db,
			Mongo:          mdb,
			Redis:          rdb,
			expectedBody:   `{"message":"OK"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test health check failure",
			DB:             &sqldb.Client{DB: sqlxMock},
			Mongo:          &mongodb.Client{Client: &mongo.Client{}},
			Redis:          &redisdb.Client{DB: redisClientMock},
			expectedBody:   `{"message":"healthcheck failed","errors":["Unable to ping postgres","Unable to ping mongo","Unable to ping redis"]}`,
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.DB, tt.Mongo, tt.Redis, "")

			req, err := http.NewRequest("GET", "/healthcheck", nil)
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			h.Check(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestHealth_Stop(t *testing.T) {
	db := sqldb.NewTestDB(t)
	mdb := mongodb.NewTestDB(t)
	rdb := redisdb.NewTestDB(t)

	h := New(db, mdb, rdb, "3000")

	testServer := httptest.NewServer(http.HandlerFunc(h.Check))
	defer testServer.Close()
	h.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/healthcheck")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)
	h.Stop()

	_, err = http.Get(testServer.URL + "/healthcheck")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/healthcheck": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}
