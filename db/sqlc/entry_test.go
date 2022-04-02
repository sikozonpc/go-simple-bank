package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateEntry(t *testing.T) {
	createRandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	newEntry := createRandomEntry(t)

	entry, err := testQueries.GetEntry(context.Background(), newEntry.ID)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, entry.AccountID, newEntry.AccountID)
	require.Equal(t, entry.Amount, newEntry.Amount)
	require.WithinDuration(t, entry.CreatedAt, newEntry.CreatedAt, time.Second)
}

func TestListEntry(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomEntry(t)
	}

	params := ListEntriesParams{
		Limit:  5,
		Offset: 5,
	}

	entries, err := testQueries.ListEntries(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}

func createRandomEntry(t *testing.T) Entry {
	acc := createRandomAccount(t)

	params := CreateEntryParams{
		AccountID: acc.ID,
		Amount:    acc.Balance / 2,
	}

	entry, err := testQueries.CreateEntry(context.Background(), params)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, entry.AccountID, params.AccountID)
	require.Equal(t, entry.Amount, params.Amount)

	return entry
}
