package db

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"practice-docker/util"
	"testing"
	"time"
)

func getFirstAccount(t *testing.T) Accounts {
	arg := GetAccountsParams{
		Limit:  1,
		Offset: 0,
	}

	accounts, err := testQueries.GetAccounts(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	return accounts[0]
}

func createRandomEntry(t *testing.T, account Accounts) Entries {

	id := sql.NullInt64{
		Int64: account.ID,
		Valid: true,
	}

	arg := CreateEntryParams{
		AccountID: id,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestQueries_CreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account)
}

func TestQueries_GetEntry(t *testing.T) {
	account := createRandomAccount(t)

	entry1 := createRandomEntry(t, account)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)
}

func TestQueries_GetEntries(t *testing.T) {
	account := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		createRandomEntry(t, account)
	}

	arg := ListEntriesParams{
		AccountID: sql.NullInt64{
			Int64: account.ID,
			Valid: true,
		},
		Limit:  5,
		Offset: 5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}
