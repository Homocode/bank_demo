package db

import (
	"context"
	"testing"
	"time"

	"github.com/homocode/bank_demo/util"
	"github.com/stretchr/testify/require"
)

func persistRandomEntry(t *testing.T, a Accounts) (Entries, CreateEntryParams, error) {
	t.Helper()

	arg := CreateEntryParams{
		AccountID: a.ID,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)

	return entry, arg, err

}

func TestCreateEntry(t *testing.T) {
	user, _, _ := persistRandomUser(t, "")
	testAccount, _, _ := persistRandomAccount(t, user, "")
	entry, arg, err := persistRandomEntry(t, testAccount)

	require.NoError(t, err)

	require.NotEmpty(t, entry)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt.Valid)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)
}

func TestGetEntry(t *testing.T) {
	user, _, _ := persistRandomUser(t, "")
	testAccount, _, _ := persistRandomAccount(t, user, "")
	entry, _, _ := persistRandomEntry(t, testAccount)

	retrievedEntry, err := testQueries.GetEntry(context.Background(), entry.ID)

	require.NoError(t, err)
	require.NotEmpty(t, retrievedEntry)

	require.Equal(t, entry.ID, retrievedEntry.ID)
	require.Equal(t, entry.AccountID, retrievedEntry.AccountID)
	require.Equal(t, entry.Amount, retrievedEntry.Amount)

	require.WithinDuration(t, entry.CreatedAt.Time, retrievedEntry.CreatedAt.Time, time.Second)
}

func TestListEntries(t *testing.T) {
	user, _, _ := persistRandomUser(t, "")
	testAccount, _, _ := persistRandomAccount(t, user, "")

	n := 10

	arg := ListEntriesParams{
		AccountID: testAccount.ID,
		Limit:     int32(n),
		Offset:    0,
	}

	for i := 0; i < n; i++ {
		persistRandomEntry(t, testAccount)
	}

	retrievedListEntries, err := testQueries.ListEntries(context.Background(), arg)

	require.NoError(t, err)

	require.NotEmpty(t, retrievedListEntries)

	require.Len(t, retrievedListEntries, n)

	for _, entry := range retrievedListEntries {
		require.NotEmpty(t, entry)
	}

}
