package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iconimpact/replaceme/internal/mongodb"
	"github.com/iconimpact/replaceme/internal/redisdb"
	"github.com/iconimpact/replaceme/internal/sqldb"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func initTestREST(t *testing.T) *REST {
	logger := zap.NewNop().Sugar()

	db, err := sqldb.New("localhost", "5432", "replaceme", "replaceme", "replaceme_test", "disable", logger)
	assert.NoError(t, err)

	mdb, err := mongodb.New("localhost", "27017", "root", "root", "mongo_db", false, logger)
	assert.NoError(t, err)

	rdb, err := redisdb.New("localhost", "6379", "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81", 15, logger)
	assert.NoError(t, err)

	return &REST{
		DB:     db,
		Mongo:  mdb,
		Redis:  rdb,
		logger: logger,
	}
}

func TestREST_GetRoot(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "TestREST_GetRoot-Success",
			body:               "",
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message":"Hello, world!"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := initTestREST(t)

			req, err := http.NewRequest("GET", "/", strings.NewReader(tt.body))
			assert.NoError(t, err)
			w := httptest.NewRecorder()
			r.GetRoot(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}
