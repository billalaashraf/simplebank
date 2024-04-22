package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/billalaashraf/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(connection)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance.Int.Int64(), account2.Balance.Int.Int64())

	n := 5
	amount := util.RandomMoney(0, 10)

	fmt.Println(">> amount:", amount.Int.Int64())

	errs := make(chan error)


	for i:=0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func () {
			ctx := context.Background()
			_, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: 	fromAccountID,
				ToAccountID: 		toAccountID,
				Amount: 				amount,
			})
			errs <- err
		}()
	}

	//check the results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

		updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
		require.NoError(t, err)
		
		updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
		require.NoError(t, err)

		fmt.Println(">> after:", updatedAccount1.Balance.Int.Int64(), updatedAccount2.Balance.Int.Int64())

		require.Equal(t, account1.Balance.Int.Int64(), updatedAccount1.Balance.Int.Int64())
		require.Equal(t, account2.Balance.Int.Int64(), updatedAccount2.Balance.Int.Int64())
}

func TestTransferTx(t *testing.T) {
	store := NewStore(connection)
	existed := make(map[int]bool)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance.Int.Int64(), account2.Balance.Int.Int64())

	n := 10
	amount := util.RandomMoney(0, 10)

	fmt.Println(">> amount:", amount.Int.Int64())

	errs := make(chan error)
	results := make(chan TransferTxResult)


	for i:=0; i < n; i++ {
		go func () {
			ctx := context.Background()
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: 	account1.ID,
				ToAccountID: 		account2.ID,
				Amount: 				amount,
			})
			errs <- err
			results <- result
		}()
	}

	//check the results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		fmt.Println(">> ReceiverAccountBalance:  ", result.ToAccount.Balance.Int.Int64())

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount.Int.Int64(), fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount.Int.Int64(), toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts 
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check accounts balance
		diff1 := account1.Balance.Int.Int64() - fromAccount.Balance.Int.Int64()
		diff2 := toAccount.Balance.Int.Int64() - account2.Balance.Int.Int64()
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount.Int.Int64() == 0)

		k := int(diff1 / amount.Int.Int64())
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

		updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
		require.NoError(t, err)
		
		updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
		require.NoError(t, err)

		fmt.Println(">> after:", updatedAccount1.Balance.Int.Int64(), updatedAccount2.Balance.Int.Int64())

		require.Equal(t, account1.Balance.Int.Int64() - int64(n) * amount.Int.Int64(), updatedAccount1.Balance.Int.Int64())
		require.Equal(t, account2.Balance.Int.Int64() + int64(n) * amount.Int.Int64(), updatedAccount2.Balance.Int.Int64())
}