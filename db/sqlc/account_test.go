package db

import (
	"context"
	"database/sql"
	"simplebank/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	newAcc := createRandomAccount(t)

	acc, err := testQueries.GetAccount(context.Background(), newAcc.ID)

	require.NoError(t, err)
	require.NotEmpty(t, acc)

	require.Equal(t, acc.ID, newAcc.ID)
	require.Equal(t, acc.Balance, newAcc.Balance)
	require.Equal(t, acc.Owner, newAcc.Owner)
	require.Equal(t, acc.Currency, newAcc.Currency)
	require.WithinDuration(t, acc.CreatedAt, newAcc.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	newAcc := createRandomAccount(t)

	params := UpdateAccountParams{
		ID:      newAcc.ID,
		Balance: util.RandomMoney(),
	}

	updatedAcc, err := testQueries.UpdateAccount(context.Background(), params)

	require.NoError(t, err)
	require.NotEmpty(t, updatedAcc)

	require.Equal(t, updatedAcc.ID, newAcc.ID)
	require.Equal(t, updatedAcc.Balance, params.Balance)
	require.Equal(t, updatedAcc.Owner, newAcc.Owner)
	require.Equal(t, updatedAcc.Currency, newAcc.Currency)
	require.WithinDuration(t, updatedAcc.CreatedAt, newAcc.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	newAcc := createRandomAccount(t)
	err := testQueries.DeleteAccount(context.Background(), newAcc.ID)

	require.NoError(t, err)
	
	acc, err := testQueries.GetAccount(context.Background(), newAcc.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, acc)
}

func TestListAccount(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	params := ListAccountsParams{
		Offset: 5,
		Limit: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, acc := range accounts {
		require.NotEmpty(t, acc)
	}
}

func createRandomAccount(t *testing.T) Account {
	params := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	acc, err := testQueries.CreateAccount(context.Background(), params)

	require.NoError(t, err)
	require.NotEmpty(t, acc)

	require.Equal(t, params.Owner, acc.Owner)
	require.Equal(t, params.Balance, acc.Balance)
	require.Equal(t, params.Currency, acc.Currency)

	require.NotZero(t, acc.ID)
	require.NotZero(t, acc.CreatedAt)

	return acc
}
