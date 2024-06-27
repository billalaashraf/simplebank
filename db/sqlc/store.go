package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute mock db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute real db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// executeTranscation executres a function within a database transaction
func (store *SQLStore) executeTransaction(ctx context.Context, fn func(*Queries) error) error {
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

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.executeTransaction(ctx, func(query *Queries) error {
		var err error
		transfer := CreateTransferParams(TransferTxParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})

		result.Transfer, err = query.CreateTransfer(ctx, transfer)
		if err != nil {
			return err
		}

		result.FromEntry, err = query.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = query.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, query, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, query, arg.ToAccountID, arg.Amount, arg.FromAccountID, arg.Amount)
		}
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}

func addMoney(
	ctx context.Context,
	queries *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64) (account1 Account, account2 Account, err error) {

	account1, err = queries.AddAccountBalance(ctx, AddAccountBalanceParams{
		AccountID: accountID1,
		Amount:    amount1,
	})
	if err != nil {
		return
	}
	account2, err = queries.AddAccountBalance(ctx, AddAccountBalanceParams{
		AccountID: accountID2,
		Amount:    amount2,
	})
	return
}
