package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store providers all functions to execute db queries and transactions
type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (s *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	// records that money is moving out
	FromEntry Entry `json:"from_entry"`
	// records that money is moving in
	ToEntry Entry `json:"to_entry"`
}

// TransferTx performes an account transfer between 2 accounts
// It creates a transfer record, add accounts entries and update accounts balance
// within a transaction
func (s *Store) TransferTx(ctx context.Context, p TransferTxParams) (TransferTxResult, error) {
	var res TransferTxResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error
		res.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: p.FromAccountID,
			ToAccountID:   p.ToAccountID,
			Amount:        p.Amount,
		})
		if err != nil {
			return err
		}

		// creating the sender entry
		res.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: p.FromAccountID,
			Amount:    -p.Amount,
		})
		if err != nil {
			return err
		}

		// creating the receiver entry
		res.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: p.ToAccountID,
			Amount:    p.Amount,
		})
		if err != nil {
			return err
		}

		if p.FromAccountID < p.ToAccountID {
			res.FromAccount, res.ToAccount, err = addMoney(ctx, q, p.FromAccountID, -p.Amount, p.ToAccountID, p.Amount)
		} else {
			res.ToAccount, res.FromAccount, err = addMoney(ctx, q, p.ToAccountID, p.Amount, p.FromAccountID, -p.Amount)
		}
		if err != nil {
			return err
		}

		return nil
	})

	return res, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accFromID int64,
	accFromAmount int64,
	accToID int64,
	accToAmount int64,
) (a1 Account, a2 Account, err error) {
	a1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accFromID,
		Amount: accFromAmount,
	})
	if err != nil {
		return
	}
	a2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accToID,
		Amount: accToAmount,
	})
	if err != nil {
		return
	}
	return
}
