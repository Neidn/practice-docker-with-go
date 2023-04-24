package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewStore creates a new store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(queries *Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfers `json:"transfer"`
	FromAccount Accounts  `json:"from_account"`
	ToAccount   Accounts  `json:"to_account"`
	FromEntry   Entries   `json:"from_entry"`
	ToEntry     Entries   `json:"to_entry"`
}

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		fromId := sql.NullInt64{Int64: arg.FromAccountID, Valid: true}
		toId := sql.NullInt64{Int64: arg.ToAccountID, Valid: true}

		// create a transfer
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: fromId,
			ToAccountID:   toId,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}

		// create from entry
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: fromId,
			Amount:    -arg.Amount,
		})

		if err != nil {
			return err
		}

		// create to entry
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: toId,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}

		// we should update the account balance in the same order
		// to avoid deadlock
		if arg.FromAccountID < arg.ToAccountID {
			// update from count to account
			result.FromAccount, result.ToAccount, err =
				addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			// update to count from account
			result.ToAccount, result.FromAccount, err =
				addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Accounts, account2 Accounts, err error) {

	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:      accountID1,
		Balance: amount1,
	})

	if err != nil {
		return Accounts{}, Accounts{}, err
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:      accountID2,
		Balance: amount2,
	})

	if err != nil {
		return Accounts{}, Accounts{}, err
	}

	return account1, account2, err
}
