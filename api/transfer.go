package api

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	db "practice-docker/db/sqlc"
)

// validAccount checks if the account exists and is owned by the given user.
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) bool {

	// check if the account exists
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		// check if the error is not ErrAccountNotFound
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	// check if the account is in the correct currency
	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true
}

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

// POST /transfers
func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if the from account exists and is owned by the user
	if !server.validAccount(ctx, req.FromAccountID, req.Currency) {
		return
	}

	// check if the to account exists and is owned by the user
	if !server.validAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	// execute the transfer in the database
	result, err := server.store.TransferTx(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// return the result to the client
	ctx.JSON(http.StatusOK, result)
}
