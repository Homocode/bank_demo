package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/homocode/bank_demo/util"
	"github.com/stretchr/testify/require"
)

func persistRandomUser(t *testing.T, owner string) (Users, CreateUserParams, error) {
	t.Helper()

	if owner == "" {
		owner = util.RandomOwner()
	}

	arg := CreateUserParams{
		Email:          owner,
		HashedPassword: "adsadads",
		FullName:       "German",
	}
	fmt.Println("user params", arg)
	user, err := testQueries.CreateUser(context.Background(), arg)

	return user, arg, err
}

func TestCreateUser(t *testing.T) {
	user, arg, err := persistRandomUser(t, "")
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.NotEmpty(t, user.CreatedAt.Valid)
	require.Equal(t, true, user.PasswordChangedAt.Time.IsZero())

	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
}

func TestGetUser(t *testing.T) {
	user, _, _ := persistRandomUser(t, "")
	retrivedUser, err := testQueries.GetUser(context.Background(), user.Email)

	require.NoError(t, err)
	require.NotEmpty(t, retrivedUser)

	require.Equal(t, user.Email, retrivedUser.Email)
	require.Equal(t, user.HashedPassword, retrivedUser.HashedPassword)
	require.Equal(t, user.FullName, retrivedUser.FullName)
	require.Equal(t, user.PasswordChangedAt, retrivedUser.PasswordChangedAt)
	require.Equal(t, user.CreatedAt, retrivedUser.CreatedAt)

	require.WithinDuration(t, user.CreatedAt.Time, retrivedUser.CreatedAt.Time, time.Second)
}
