package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "practice-docker/db/sqlc"
	"practice-docker/token"
	"practice-docker/util"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// Set up the routing of the server.
func (server *Server) setupRouter() {
	router := gin.Default()

	// Set up the mode of the server.
	gin.SetMode(gin.DebugMode)
	err := router.SetTrustedProxies([]string{
		"127.0.0.1",
	})
	if err != nil {
		return
	}

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	router.POST("/transfers", server.createTransfer)

	server.router = router
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, err
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	// Register the custom validator.
	if validate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := validate.RegisterValidation("currency", validCurrency)
		if err != nil {
			return nil, err
		}
	}

	// Set up the routing of the server.
	// START //
	server.setupRouter()
	// END //

	return server, nil
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// errorResponse handles the error response.
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
