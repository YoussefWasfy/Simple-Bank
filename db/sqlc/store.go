package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TranferTxResult, error)
}

// Provide all functions to execute all db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// executes a function within a database transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

// contains the input parameters to of the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// Contains the result of the transfer transaction
type TranferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	fromAccount Account  `json:"from_account"`
	toAccount   Account  `json:"to_account"`
	fromEntry   Entry    `json:"from_entry"`
	toEntry     Entry    `json:"to_entry"`
}

// Performs a money transfer from one account to another
// It creates a transfer record, add account entries, and update account balance within a single database transaction
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TranferTxResult, error) {
	var result TranferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}
		result.fromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}
		result.toEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}
		if arg.FromAccountID < arg.ToAccountID {

			result.fromAccount, result.toAccount, err = AddMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.toAccount, result.fromAccount, err = AddMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return nil
	})

	return result, err
}

func AddMoney(ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2,
	amount2 int64) (account1 Account, account2 Account, err error) {

	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}
	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}
