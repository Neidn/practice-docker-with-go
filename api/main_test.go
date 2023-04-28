package api

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	db "practice-docker/db/sqlc"
	"practice-docker/util"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenLifetime: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoErrorf(t, err, "failed to create server: %v", err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
