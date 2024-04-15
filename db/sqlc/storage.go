package db

import (
	"context"
	"fmt"

	"github.com/billalaashraf/simplebank/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

//Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries
	db *pgx.Conn
}

type TransferTxParams struct {
	FromAccountID int64 					`json:"from_account_id"`
	ToAccountID 	int64 					`json:"to_account_id"`
	Amount 				pgtype.Numeric 	`json:"amount"`
}

type TransferTxResult struct {
	Transfer Transfer `json:"transfer"`
	FromAccount Account `json:"from_account"`
	ToAccount Account `json:"to_account"`
	FromEntry Entry `json:"from_entry"`
	ToEntry Entry `json:"to_entry"`
}

func NewStore(db *pgx.Conn) *Store {
		return &Store{
			db: db,
			Queries: New(db),
		}
}
//executeTranscation executres a function within a database transaction
func (store *Store) executeTransaction(ctx context.Context, fn func(*Queries) error) error {
	options := pgx.TxOptions {}
	transaction, err := store.db.BeginTx(ctx, options)
	if err != nil {
		return err
	}
	query := New(transaction)
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
		transfer := CreateTransferParams(TransferTxParams {
			FromAccountID: 	arg.FromAccountID,
			ToAccountID: 		arg.ToAccountID,
			Amount: 				arg.Amount,
		})
		

		result.Transfer, err = query.CreateTransfer(ctx, transfer)
		if err != nil {
			return err
		}

		result.FromEntry, err = query.CreateEntry(ctx, CreateEntryParams {
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount.Int.Int64(),
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = query.CreateEntry(ctx, CreateEntryParams {
			AccountID: arg.ToAccountID,
			Amount: arg.Amount.Int.Int64(),
		})
		if err != nil {
			return err
		}

		// TODO: update account's balance
		account1, err := query.GetAccountForUpdate(ctx, arg.FromAccountID);
		if err != nil {
			return err
		}

		result.FromAccount, err = query.UpdateAccount(ctx, UpdateAccountParams {
			ID: arg.FromAccountID,
			Balance: util.FromIntToPgNumeric(account1.Balance.Int.Int64() - arg.Amount.Int.Int64()),
		})
		if err != nil {
			return err
		}

		account2, err := query.GetAccountForUpdate(ctx, arg.ToAccountID);
		if err != nil {
			return err
		}

		result.ToAccount, err = query.UpdateAccount(ctx, UpdateAccountParams {
			ID: arg.ToAccountID,
			Balance: util.FromIntToPgNumeric(account2.Balance.Int.Int64() + arg.Amount.Int.Int64()),
		})

		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}