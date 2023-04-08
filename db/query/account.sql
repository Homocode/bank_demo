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

-- name: ListAccounts :many
SELECT * FROM accounts
WHERE owner = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: AddAmountToAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount) -- equal to: balance + $1
WHERE id = sqlc.arg(id) -- equal to: $2
RETURNING *;
