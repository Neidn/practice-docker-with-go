package db

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"practice-docker/util"
	"testing"
	"time"
)

func getFromAccountID(t *testing.T) sql.NullInt64 {
	args := GetAccountsParams{
		Limit:  1,
		Offset: 0,
	}

	accounts, err := testQueries.GetAccounts(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	return sql.NullInt64{
		Int64: accounts[0].ID,
		Valid: true,
	}
}

func getToAccountID(t *testing.T) sql.NullInt64 {
	args := GetAccountsParams{
		Limit:  1,
		Offset: 1,
	}

	accounts, err := testQueries.GetAccounts(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	return sql.NullInt64{
		Int64: accounts[0].ID,
		Valid: true,
	}
}

func createRandomTransfer(t *testing.T, fromID sql.NullInt64, toID sql.NullInt64) Transfers {
	arg := CreateTransferParams{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        util.RandomMoney(),
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID.Int64, transfer.FromAccountID.Int64)
	require.Equal(t, arg.ToAccountID.Int64, transfer.ToAccountID.Int64)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestQueries_CreateTransfer(t *testing.T) {
	fromID := getFromAccountID(t)
	toID := getToAccountID(t)

	createRandomTransfer(t, fromID, toID)
}

func TestQueries_GetTransfer(t *testing.T) {
	fromID := getFromAccountID(t)
	toID := getToAccountID(t)

	transfer := createRandomTransfer(t, fromID, toID)

	transfer2, err := testQueries.GetTransfer(context.Background(), transfer.ID)

	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer.ID, transfer2.ID)
	require.Equal(t, transfer.FromAccountID.Int64, transfer2.FromAccountID.Int64)
	require.Equal(t, transfer.ToAccountID.Int64, transfer2.ToAccountID.Int64)
	require.Equal(t, transfer.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer.CreatedAt, transfer2.CreatedAt, time.Second)
}

func TestQueries_ListTransfers(t *testing.T) {
	fromID := getFromAccountID(t)
	toID := getToAccountID(t)

	for i := 0; i < 10; i++ {
		createRandomTransfer(t, fromID, toID)
	}

	arg := ListTransfersParams{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Limit:         5,
		Offset:        5,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}
