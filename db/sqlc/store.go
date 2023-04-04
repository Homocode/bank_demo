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
		if arg.FromAccountId < arg.ToAccountId {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountId, -arg.Amount, arg.ToAccountId, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountId, arg.Amount, arg.FromAccountId, -arg.Amount)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Accounts, account2 Accounts, err error) {
	account1, err = q.AddAmountToAccountBalance(ctx, AddAmountToAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAmountToAccountBalance(ctx, AddAmountToAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})

	return
}
