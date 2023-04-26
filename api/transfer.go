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

type getTransferRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// GET /transfers/:id
func (server *Server) getTransfer(ctx *gin.Context) {
	var req getTransferRequest
	err := ctx.ShouldBindUri(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get the transfer from the database
	transfer, err := server.store.GetTransfer(ctx, req.ID)
	if err != nil {
		// check if the error is not ErrTransferNotFound
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// return the transfer to the client
	ctx.JSON(http.StatusOK, transfer)
}

type listTransfersRequest struct {
	FromAccountID int64  `form:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `form:"to_account_id" binding:"required,min=1"`
	FromCurrency  string `form:"from_currency" binding:"required,currency"`
	ToCurrency    string `form:"to_currency" binding:"required,currency"`
	PageID        int32  `form:"page_id" binding:"required,min=1"`
	PageSize      int32  `form:"page_size" binding:"required,min=1,max=100"`
}

// GET /transfers
func (server *Server) listTransfers(ctx *gin.Context) {
	var req listTransfersRequest
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// check if the from account exists and is owned by the user
	if !server.validAccount(ctx, req.FromAccountID, req.FromCurrency) {
		return
	}

	// check if the to account exists and is owned by the user
	if !server.validAccount(ctx, req.ToAccountID, req.ToCurrency) {
		return
	}

	reqFromAccountID := sql.NullInt64{
		Int64: req.FromAccountID,
		Valid: true,
	}

	reqToAccountID := sql.NullInt64{
		Int64: req.ToAccountID,
		Valid: true,
	}

	// get the transfers from the database
	arg := db.ListTransfersParams{
		FromAccountID: reqFromAccountID,
		ToAccountID:   reqToAccountID,
		Limit:         req.PageSize,
		Offset:        (req.PageID - 1) * req.PageSize,
	}

	transfers, err := server.store.ListTransfers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// return the transfers to the client
	ctx.JSON(http.StatusOK, transfers)
}
