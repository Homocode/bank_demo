-- name: CreateAccount :one
INSERT INTO accounts (
    owner,
    balance,
    currency
) VALUES (
    $1, $2, $3
) 
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: GetAccountToUpdate :one
SELECT * FROM accounts
WHERE id = $1 
LIMIT 1
FOR NO KEY UPDATE
-- FOR UPDATE is used in a Tx (let's call it Tx1) to lock the table for other Tx's, thus, no other queries 
-- coming from another Tx (let's call it Tx2) can be executed on the table, so if Tx2 attemps to operate 
-- on the lock table, Tx2 gets block until Tx1 (that performs the lock) is finished (COMMITED or ROLLBACK).
-- Only then Tx2 is able to operate on accounts table and the Tx is "unblock".
--
-- NO KEY tell postgres that the primary key isnt going to be update.
-- If NO KEY isnt use a "DEADLOCK" will occur because other tables use the accounts table
-- primary key (id) as their foregein keys.
-- When the table accounts get lock (because of the use of FOR UPDATE) from Tx1, Tx2 can´t operate 
-- (read/write) from the accounts table row that is lock, so when Tx2 need to access the account row to
-- get the account id to use it as a foregin key for a new record on another table (transfers or entries)
-- it can't access it due to the lock that is being held on accounts table row from Tx1. So Tx2 also get
-- block and that is why the deadlock occurs (a deadlock is a situation in which two or more transactions 
-- are waiting for one another to give up locks.)
;

-- name: ListAccounts :many
SELECT * FROM accounts
WHERE owner = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;

-- name: AddAmountToAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount) -- equal to: balance + $1
WHERE id = sqlc.arg(id) -- equal to: $2
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;