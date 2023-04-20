package db

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDb)

	user1, _, _ := persistRandomUser(t, "")
	account1, _, _ := persistRandomAccount(t, user1, "")
	user2, _, _ := persistRandomUser(t, "")
	account2, _, _ := persistRandomAccount(t, user2, "")
	amount := 400

	arg := TransferTxParams{
		FromAccountId: account1.ID,
		ToAccountId:   account2.ID,
		Amount:        int64(amount),
	}

	// Run n concurrent TransferTx. Each one in one of
	// the n go rutines.
	n := 5
	results := make(chan TransferTxResult, n)
	errors := make(chan error, n)
	var wg = &sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {

		go func() {
			defer wg.Done()
			result, err := store.TransferTx(context.Background(), arg)
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

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDb)

	user1, _, _ := persistRandomUser(t, "")
	account1, _, _ := persistRandomAccount(t, user1, "")
	user2, _, _ := persistRandomUser(t, "")
	account2, _, _ := persistRandomAccount(t, user2, "")
	amount := 400

	// Run n concurrent TransferTx. Each one in one of
	// the n go rutines.
	n := 10
	errors := make(chan error, n)
	var wg = &sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		arg := TransferTxParams{
			FromAccountId: account1.ID,
			ToAccountId:   account2.ID,
			Amount:        int64(amount),
		}

		// for half of go rutines the Tx is from acc2 to acc1
		if i%2 == 0 {
			arg = TransferTxParams{
				FromAccountId: account2.ID,
				ToAccountId:   account1.ID,
				Amount:        int64(amount),
			}
		}

		go func() {
			defer wg.Done()
			_, err := store.TransferTx(context.Background(), arg)
			errors <- err
		}()
	}
	wg.Wait()

	// Validate test cases
	for i := 0; i < n; i++ {
		err := <-errors
		require.NoError(t, err)
	}

	// check persisted final balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	// As half of the transactions are going from acc1 to acc2 and the other half viceversa,
	// the fianl balance is the same as the intial
	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
