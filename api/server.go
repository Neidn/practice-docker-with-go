package api

import (
	"github.com/gin-gonic/gin"
	db "practice-docker/db/sqlc"
)

type Server struct {
	store  *db.SQLStore
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store *db.SQLStore) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Set up the routing of the server.
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	server.router = router
	return server
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// errorResponse handles the error response.
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
