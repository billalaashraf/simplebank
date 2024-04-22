package db

import (
	"context"
	"fmt"

	"github.com/billalaashraf/simplebank/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries
	db *pgxpool.Pool
}

type TransferTxParams struct {
	FromAccountID int64          `json:"from_account_id"`
	ToAccountID   int64          `json:"to_account_id"`
	Amount        pgtype.Numeric `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// executeTranscation executres a function within a database transaction
func (store *Store) executeTransaction(ctx context.Context, fn func(*Queries) error) error {
	options := pgx.TxOptions{}
	transaction, err := store.db.BeginTx(ctx, options)
	if err != nil {
		return err
	}
	query := store.WithTx(transaction)
	err = fn(query)
	if err != nil {
		if rollbackError := transaction.Rollback(ctx); rollbackError != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rollbackError)
		}
		return err
	}
	return transaction.Commit(ctx)
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
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
			Amount:    -arg.Amount.Int.Int64(),
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = query.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount.Int.Int64(),
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, query, arg.FromAccountID, -arg.Amount.Int.Int64(), arg.ToAccountID, arg.Amount.Int.Int64())
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, query, arg.ToAccountID, arg.Amount.Int.Int64(), arg.FromAccountID, arg.Amount.Int.Int64())
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
		Amount:    util.FromIntToPgNumeric(amount1),
	})
	if err != nil {
		return
	}
	account2, err = queries.AddAccountBalance(ctx, AddAccountBalanceParams{
		AccountID: accountID2,
		Amount:    util.FromIntToPgNumeric(amount2),
	})
	return
}
