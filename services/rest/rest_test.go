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

func TestREST_StartStop(t *testing.T) {
	logger := zap.NewNop().Sugar()
	db, err := sqldb.New("localhost", "5432", "replaceme", "replaceme", "replaceme_test", "disable", logger)
	assert.NoError(t, err)

	mdb, err := mongodb.New("localhost", "27017", "root", "root", "mongo_db", false, logger)
	assert.NoError(t, err)

	rdb, err := redisdb.New("localhost", "6379", "eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81", 15, logger)
	assert.NoError(t, err)
	h := New(db, mdb, rdb, "", logger)

	testServer := httptest.NewServer(http.HandlerFunc(h.GetRoot))
	defer testServer.Close()
	h.srv = testServer.Config
	go h.Start()

	resp, err := http.Get(testServer.URL + "/")
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK)
	h.Stop()

	_, err = http.Get(testServer.URL + "/")
	assert.Error(t, err)
	assert.Equal(t, `Get "`+testServer.URL+`/": dial tcp 127.0.0.1:`+strings.Split(testServer.URL, ":")[2]+`: connect: connection refused`, err.Error())
}
