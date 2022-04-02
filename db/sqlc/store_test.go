package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	s := NewStore(testDB)

	fromAcc := createRandomAccount(t)
	toAcc := createRandomAccount(t)
	fmt.Println(">> before tx: ", fromAcc.Balance, toAcc.Balance)

	// test concurrent transfer transactions
	n := 5
	amount := int64(10)

	goroutineErrsChan := make(chan error)
	txResChan := make(chan TransferTxResult)

	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		go func() {
			res, err := s.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAcc.ID,
				ToAccountID:   toAcc.ID,
				Amount:        amount,
			})

			goroutineErrsChan <- err
			txResChan <- res
		}()
	}

	for i := 0; i < n; i++ {
		err := <-goroutineErrsChan
		require.NoError(t, err)

		res := <-txResChan
		require.NotEmpty(t, res)

		// check the transfer
		transfer := res.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.FromAccountID, fromAcc.ID)
		require.Equal(t, transfer.ToAccountID, toAcc.ID)
		require.Equal(t, transfer.Amount, amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// check that a record in the DB was actually created
		_, err = s.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := res.FromEntry
		_, err = s.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		require.Equal(t, fromEntry.AccountID, fromAcc.ID)
		require.Equal(t, fromEntry.Amount, -amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		toEntry := res.ToEntry
		_, err = s.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		require.Equal(t, toEntry.AccountID, toAcc.ID)
		require.Equal(t, toEntry.Amount, amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		// check accounts
		fromAccount := res.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, fromAcc.ID)

		toAccount := res.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, toAcc.ID)

		// check accounts balance
		fmt.Println(">> tx: ", fromAccount.Balance, toAccount.Balance)

		lostAmmount := fromAcc.Balance - fromAccount.Balance
		receivedAmmount := toAccount.Balance - toAcc.Balance
		require.Equal(t, lostAmmount, receivedAmmount)
		require.True(t, lostAmmount > 0)
		require.True(t, lostAmmount%amount == 0)

		k := int(lostAmmount / amount)
		require.True(t, k >= 1 && k <= 5)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balances
	updatedFromAccount, err := testQueries.GetAccount(context.Background(), fromAcc.ID)
	require.NoError(t, err)

	updatedToAccount, err := testQueries.GetAccount(context.Background(), toAcc.ID)
	require.NoError(t, err)

	fmt.Println(">> after tx: ", updatedFromAccount.Balance, updatedToAccount.Balance)

	require.Equal(t, fromAcc.Balance-int64(n)*amount, updatedFromAccount.Balance)
	require.Equal(t, updatedToAccount.Balance, toAcc.Balance+int64(n)*amount)
}

func TestTransferTxDeadblock(t *testing.T) {
	s := NewStore(testDB)

	fromAcc := createRandomAccount(t)
	toAcc := createRandomAccount(t)
	fmt.Println(">> before tx: ", fromAcc.Balance, toAcc.Balance)

	// test concurrent transfer transactions
	n := 10
	amount := int64(10)
	goroutineErrsChan := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := fromAcc.ID
		toAccountID := toAcc.ID

		// to make 50% send, we're checking if the index is odd
		if i%2 == 1 {
			fromAccountID = toAcc.ID
			toAccountID = fromAcc.ID
		}

		go func() {
			_, err := s.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			goroutineErrsChan <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-goroutineErrsChan
		require.NoError(t, err)
	}

	// check the final updated balances
	updatedFromAccount, err := testQueries.GetAccount(context.Background(), fromAcc.ID)
	require.NoError(t, err)

	updatedToAccount, err := testQueries.GetAccount(context.Background(), toAcc.ID)
	require.NoError(t, err)

	fmt.Println(">> after tx: ", updatedFromAccount.Balance, updatedToAccount.Balance)

	require.Equal(t, fromAcc.Balance, updatedFromAccount.Balance)
	require.Equal(t, updatedToAccount.Balance, toAcc.Balance)
}
