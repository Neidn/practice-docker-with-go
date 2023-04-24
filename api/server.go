package api

import (
	"github.com/gin-gonic/gin"
	db "practice-docker/db/sqlc"
)

type Server struct {
	store  *db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Set up the routing of the server.
	router.POST("/accounts", server.createAccount)

	server.router = router
	return server
}
