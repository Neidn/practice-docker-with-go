package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "practice-docker/db/sqlc"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Register the custom validator.
	if validate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := validate.RegisterValidation("currency", validCurrency)
		if err != nil {
			return nil
		}
	}

	// Set up the mode of the server.
	gin.SetMode(gin.DebugMode)
	err := router.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return nil
	}

	// Set up the routing of the server.
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	router.POST("/transfers", server.createTransfer)

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
