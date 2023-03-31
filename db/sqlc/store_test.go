package db

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDb)

	account1, _, _ := persistRandomAccount(t, "")
	account2, _, _ := persistRandomAccount(t, "")
	amount := 400

	arg := TransferTxParams{
		FromAccountId: account1.ID,
		ToAccountId:   account2.ID,
		Amount:        int64(amount),
	}

	// Run n concurrent TransferTx. Each one in one of
	// the n go rutines.
	n := 2
	results := make(chan TransferTxResult, n)
	errors := make(chan error, n)
	var wg = &sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			defer wg.Done()
			ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(ctx, arg)
			results <- result
			errors <- err
		}()
	}
	wg.Wait()

	// Validate test cases
	var k int
	for i := 0; i < n; i++ {
		err := <-errors
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		//check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, arg.ToAccountId, transfer.ToAccountID)
		require.Equal(t, arg.FromAccountId, transfer.FromAccountID)
		require.Equal(t, arg.Amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt.Valid)

		// This is to check that the row persisted in the transaction
		// is available to read.
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, arg.FromAccountId, fromEntry.AccountID)
		require.Equal(t, -arg.Amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt.Valid)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, arg.ToAccountId, toEntry.AccountID)
		require.Equal(t, arg.Amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt.Valid)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, arg.FromAccountId, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, arg.ToAccountId, toAccount.ID)

		// check accounts balance
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := account2.Balance - toAccount.Balance
		require.True(t, diff1 > 0)
		require.True(t, diff2 < 0)
		require.True(t, diff1+diff2 == 0)

		// check persisted final balance
		updatedAccount1, err := testQueries.GetAccount(context.Background(), arg.FromAccountId)
		require.NoError(t, err)

		updatedAccount2, err := testQueries.GetAccount(context.Background(), arg.ToAccountId)
		require.NoError(t, err)

		k++
		txAmountAcumm := amount * k
		require.Equal(t, account1.Balance-int64(txAmountAcumm), updatedAccount1.Balance+int64(amount*(n-i-1)))
		require.Equal(t, account2.Balance+int64(txAmountAcumm), updatedAccount2.Balance-int64(amount*(n-i-1)))
	}
}
