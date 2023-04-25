package api

import (
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http"
	db "practice-docker/db/sqlc"
	"practice-docker/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required, alphanum"`
	Password string `json:"password" binding:"required, min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required, email"`
}

// POST /users
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Hash the user's password.
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	// Create a new user.
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		// Check if the error is pq.Error.
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Return the user.
	ctx.JSON(http.StatusOK, user)
}
