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
<<<<<<< HEAD
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)
=======
>>>>>>> origin/main

	server.router = router
	return server
}
<<<<<<< HEAD

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// errorResponse is a helper function to convert an error to a JSON response.
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
=======
>>>>>>> origin/main
