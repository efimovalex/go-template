package healthcheck

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/efimovalex/replaceme/adapters/mongodb"
	"github.com/efimovalex/replaceme/adapters/postgres"
	"github.com/efimovalex/replaceme/adapters/redisdb"
	"github.com/go-redis/redismock/v8"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestHealth_Check(t *testing.T) {
	db := postgres.NewTestDB(t)

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
		DB             *postgres.Client
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
			DB:             &postgres.Client{DB: sqlxMock},
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

func TestHealth_StopGrecefully(t *testing.T) {
	db := postgres.NewTestDB(t)
	mdb := mongodb.NewTestDB(t)
	rdb := redisdb.NewTestDB(t)

	h := New(db, mdb, rdb, "3000")

	testServer := httptest.NewServer(http.HandlerFunc(h.Check))
	defer testServer.Close()
	h.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/healthcheck")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	<-ctx.Done()
	time.Sleep(time.Millisecond * 5)
	h.Stop(ctx)

	_, err = http.Get(testServer.URL + "/healthcheck")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/healthcheck": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}

func TestHealth_StopDeadline(t *testing.T) {
	db := postgres.NewTestDB(t)
	mdb := mongodb.NewTestDB(t)
	rdb := redisdb.NewTestDB(t)

	h := New(db, mdb, rdb, "3000")

	testServer := httptest.NewServer(http.HandlerFunc(h.Check))
	defer testServer.Close()
	h.srv = testServer.Config

	resp, err := http.Get(testServer.URL + "/healthcheck")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(time.Millisecond * 5)
	<-ctx.Done()
	h.Stop(ctx)

	_, err = http.Get(testServer.URL + "/healthcheck")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/healthcheck": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}

func TestHealth_JSON(t *testing.T) {
	db := postgres.NewTestDB(t)
	mdb := mongodb.NewTestDB(t)
	rdb := redisdb.NewTestDB(t)

	h := New(db, mdb, rdb, "3000")

	type args struct {
		w      http.ResponseWriter
		status int
		v      interface{}
	}
	tests := []struct {
		name   string
		args   args
		status int
		want   string
	}{
		{
			name: "Test success",
			args: args{
				w:      httptest.NewRecorder(),
				status: http.StatusOK,
				v:      map[string]string{"message": "OK"},
			},
			status: http.StatusOK,
			want:   `{"message":"OK"}`,
		},
		{
			name: "Test failure",
			args: args{
				w:      httptest.NewRecorder(),
				status: http.StatusOK,
				v: map[string]float64{
					"error": math.Inf(1),
				},
			},
			status: http.StatusInternalServerError,
			want:   `{"errors":["Internal Server Error"]}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.JSON(w, tt.args.status, tt.args.v)
			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, tt.want, w.Body.String())
		})
	}
}

func TestHealth_Start(t *testing.T) {
	db := postgres.NewTestDB(t)
	mdb := mongodb.NewTestDB(t)
	rdb := redisdb.NewTestDB(t)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Log(testServer.URL)
	h := New(db, mdb, rdb, "not_a_port")

	err := h.Start(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "listen tcp: lookup tcp/not_a_port: nodename nor servname provided, or not known", err.Error())
}
