package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all the functions to execute transactions
// and db queries
type Store struct {
	//With this, Store has the properties of type Queries.
	//It`s call composition and is used to inherit functionality.
	*Queries
	db *sql.DB
}

// NewStore instantiate a new Store
func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes a function within a database transaction
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountId int64
	ToAccountId   int64
	Amount        int64
}

type TransferTxResult struct {
	Transfer    Transfers
	FromAccount Accounts
	ToAccount   Accounts
	FromEntry   Entries
	ToEntry     Entries
}

var txKey = struct{}{}

// TransferTx performs a transfer between two accounts by creating a transfer record,
// two entry records (money out FromAccount and money in ToAccount) and update accounts balance.
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		//Createa a transfer record to persist the amount and accounts involved
		//in the transference
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountId,
			ToAccountID:   arg.ToAccountId,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		//Creates an entry record to persist the amount of money leaving the account
		//where it is transferred from
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountId,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		//Creates an entry record to persist the amount of money entering the account
		//where it is transferred to
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountId,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		//Update accounts balance
		result.FromAccount, err = q.AddAmountToAccountBalance(ctx, AddAmountToAccountBalanceParams{
			ID:     arg.FromAccountId,
			Amount: -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToAccount, err = q.AddAmountToAccountBalance(ctx, AddAmountToAccountBalanceParams{
			ID:     arg.ToAccountId,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
