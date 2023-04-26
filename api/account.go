package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http"
	db "practice-docker/db/sqlc"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

// POST /accounts
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	err := ctx.ShouldBindJSON(&req)

	// If the request body is not a valid JSON or the JSON does not match the struct, return a 400 Bad Request.
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountsParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}

	// Create a new account in the database.
	account, err := server.store.CreateAccounts(ctx, arg)

	if err != nil {
		// If the account is not found, return a 404 Not Found.
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation": // error code 23503, 23505
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// GET /accounts/:id
func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	err := ctx.ShouldBindUri(&req)

	// If the request body is not a valid JSON or the JSON does not match the struct, return a 400 Bad Request.
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Get the account from the database.
	account, err := server.store.GetAccount(ctx, req.ID)

	if err != nil {
		// If the account is not found, return a 404 Not Found.
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=1,max=100"`
}

// GET /accounts
func (server *Server) listAccounts(ctx *gin.Context) {
	var req listAccountsRequest
	err := ctx.ShouldBindQuery(&req)

	// If the request body is not a valid JSON or the JSON does not match the struct, return a 400 Bad Request.
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	// Get the accounts from the database.
	accounts, err := server.store.GetAccounts(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}
