package db

import (
	"context"
	"testing"
	"time"

	"github.com/homocode/bank_demo/util"
	"github.com/stretchr/testify/require"
)

// Generetes random args and persist the account. Param owner set to "", generates random owner.
// To create the account with a specific owner set it to <name> (this is used for TestListAccounts)
func persistRandomAccount(t *testing.T, user Users) (Accounts, CreateAccountParams, error) {
	t.Helper() //Helper marks the calling function as a test helper function.

	arg := CreateAccountParams{
		Owner:    user.Email,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)

	return account, arg, err
}
func TestCreateAccount(t *testing.T) {
	user, _, _ := persistRandomUser(t, "")
	account, arg, err := persistRandomAccount(t, user)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt.Valid)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)
}

func TestGetAccount(t *testing.T) {
	user, _, _ := persistRandomUser(t, "")
	account, _, _ := persistRandomAccount(t, user)
	retrievedAccount, err := testQueries.GetAccount(context.Background(), account.ID)

	require.NoError(t, err)
	require.NotEmpty(t, retrievedAccount)

	require.Equal(t, account.ID, retrievedAccount.ID)
	require.Equal(t, account.Owner, retrievedAccount.Owner)
	require.Equal(t, account.Balance, retrievedAccount.Balance)
	require.Equal(t, account.Currency, retrievedAccount.Currency)
	require.Equal(t, account.CreatedAt, retrievedAccount.CreatedAt)

	require.WithinDuration(t, account.CreatedAt.Time, retrievedAccount.CreatedAt.Time, time.Second)
}

func TestListAccounts(t *testing.T) {
	user, userArgs, _ := persistRandomUser(t, "ger@gmail.com")
	n := 3

	arg := ListAccountsParams{
		Owner:  userArgs.Email,
		Limit:  int32(n),
		Offset: 0,
	}

	for i := 0; i < n; i++ {
		persistRandomAccount(t, user)
	}

	retrieveAccounts, err := testQueries.ListAccounts(context.Background(), arg)

	require.NoError(t, err)
	require.Len(t, retrieveAccounts, n)

	for _, account := range retrieveAccounts {
		require.NotEmpty(t, account)
	}
}
