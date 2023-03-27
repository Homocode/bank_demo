package db

import (
	"context"
	"testing"
	"time"

	"github.com/homocode/bank_demo/util"
	"github.com/stretchr/testify/require"
)

func persistRandomTransfer(t *testing.T, account1 Accounts, account2 Accounts) (Transfers, CreateTransferParams, error) {
	t.Helper()

	arg := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomMoney(),
	}

	transfer, err := testQueries.CreateTransfer(context.Background(), arg)

	return transfer, arg, err
}

func TestCreateTransfer(t *testing.T) {
	account1, _, _ := persistRandomAccount(t, "")
	account2, _, _ := persistRandomAccount(t, "")

	transfer, arg, err := persistRandomTransfer(t, account1, account2)

	require.NoError(t, err)

	require.NotEmpty(t, transfer)
	require.NotEmpty(t, transfer.ID)

	require.NotZero(t, transfer.CreatedAt.Valid)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
}

func TestGetTransfer(t *testing.T) {
	account1, _, _ := persistRandomAccount(t, "")
	account2, _, _ := persistRandomAccount(t, "")

	transfer, _, _ := persistRandomTransfer(t, account1, account2)

	retrievedTransfer, err := testQueries.GetTransfer(context.Background(), transfer.ID)

	require.NoError(t, err)

	require.NotEmpty(t, retrievedTransfer)

	require.Equal(t, transfer.FromAccountID, retrievedTransfer.FromAccountID)
	require.Equal(t, transfer.ToAccountID, retrievedTransfer.ToAccountID)
	require.Equal(t, transfer.Amount, retrievedTransfer.Amount)

	require.WithinDuration(t, transfer.CreatedAt.Time, retrievedTransfer.CreatedAt.Time, time.Second)
}

func TestListTransfer(t *testing.T) {
	account1, _, _ := persistRandomAccount(t, "")
	account2, _, _ := persistRandomAccount(t, "")

	var fromAccountId int64
	n := 10

	for i := 0; i < n; i++ {
		transfer, _, _ := persistRandomTransfer(t, account1, account2)
		fromAccountId = transfer.FromAccountID
	}

	arg := ListTransfersParams{
		FromAccountID: fromAccountId,
		Limit:         int32(n),
		Offset:        0,
	}

	retrievedListTransfer, err := testQueries.ListTransfers(context.Background(), arg)

	require.NoError(t, err)

	require.Len(t, retrievedListTransfer, n)

	for _, transfer := range retrievedListTransfer {
		require.NotEmpty(t, transfer)
	}
}
