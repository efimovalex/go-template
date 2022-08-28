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
	mongodb "github.com/efimovalex/replaceme/adapters/mongodb/mock"
	"github.com/efimovalex/replaceme/adapters/postgres"
	"github.com/efimovalex/replaceme/adapters/redisdb"
	"github.com/go-redis/redismock/v8"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestHealth_Check(t *testing.T) {
	t.Run("Test health check", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		defer mockDB.Close()
		mock.ExpectPing()
		sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

		redisClientMock, redismock := redismock.NewClientMock()
		redismock.ExpectPing().SetVal("ok")
		defer redisClientMock.Close()

		mongoClientMock := mongodb.ClientMock{}
		mongoClientMock.On("Ping").Return(nil)

		h := New(&postgres.Client{DB: sqlxMock}, &mongoClientMock, &redisdb.Client{DB: redisClientMock}, "")

		req, err := http.NewRequest("GET", "/healthcheck", nil)
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		h.Check(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"message":"OK"}`, w.Body.String())
	})

	t.Run("Test health check failure with error", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		assert.NoError(t, err)
		mock.ExpectPing().WillReturnError(errors.New("ping error"))
		sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

		defer mockDB.Close()

		redisClientMock, redismock := redismock.NewClientMock()
		redismock.ExpectPing().SetErr(errors.New("redis error"))
		defer redisClientMock.Close()

		mongoClientMock := mongodb.ClientMock{}
		mongoClientMock.On("Ping").Return(errors.New("mongo error"))

		h := New(&postgres.Client{DB: sqlxMock}, &mongoClientMock, &redisdb.Client{DB: redisClientMock}, "")

		req, err := http.NewRequest("GET", "/healthcheck", nil)
		assert.NoError(t, err)
		w := httptest.NewRecorder()
		h.Check(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, `{"message":"healthcheck failed","errors":["Unable to ping postgres","Unable to ping mongo","Unable to ping redis"]}`, w.Body.String())
	})

}

func TestHealth_StopGrecefully(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	mock.ExpectPing()
	sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

	defer mockDB.Close()

	redisClientMock, redismock := redismock.NewClientMock()
	redismock.ExpectPing().SetVal("ok")
	defer redisClientMock.Close()

	mongoClientMock := mongodb.ClientMock{}
	mongoClientMock.On("Ping").Return(nil)

	h := New(&postgres.Client{DB: sqlxMock}, &mongoClientMock, &redisdb.Client{DB: redisClientMock}, "")

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
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	mock.ExpectPing()
	sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

	defer mockDB.Close()

	redisClientMock, redismock := redismock.NewClientMock()
	redismock.ExpectPing().SetVal("ok")
	defer redisClientMock.Close()

	mongoClientMock := mongodb.ClientMock{}
	mongoClientMock.On("Ping").Return(nil)

	h := New(&postgres.Client{DB: sqlxMock}, &mongoClientMock, &redisdb.Client{DB: redisClientMock}, "")

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
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	mock.ExpectPing().WillReturnError(errors.New("ping error"))
	sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

	defer mockDB.Close()

	redisClientMock, redismock := redismock.NewClientMock()
	redismock.ExpectPing().SetErr(errors.New("redis error"))
	defer redisClientMock.Close()

	mongoClientMock := mongodb.ClientMock{}
	mongoClientMock.On("Ping").Return(nil)

	h := New(&postgres.Client{DB: sqlxMock}, &mongoClientMock, &redisdb.Client{DB: redisClientMock}, "")

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
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	mock.ExpectPing().WillReturnError(errors.New("ping error"))
	sqlxMock := sqlx.NewDb(mockDB, "sqlmock")

	defer mockDB.Close()

	redisClientMock, redismock := redismock.NewClientMock()
	redismock.ExpectPing().SetErr(errors.New("redis error"))
	defer redisClientMock.Close()

	mongoClientMock := mongodb.ClientMock{}
	mongoClientMock.On("Ping").Return(nil)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Log(testServer.URL)
	h := New(&postgres.Client{DB: sqlxMock}, &mongoClientMock, &redisdb.Client{DB: redisClientMock}, "not:a:port")

	err = h.Start(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "listen tcp: address :not:a:port: too many colons in address", err.Error())
}
