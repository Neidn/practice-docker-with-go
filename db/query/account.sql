-- name: CreateAccounts :one
INSERT INTO accounts (owner, balance, currency)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetAccount :one
SELECT *
FROM accounts
WHERE id = $1 LIMIT 1;

-- name: GetAccounts :many
SELECT *
FROM accounts
ORDER BY id LIMIT $1
OFFSET $2;

-- name: GetAccountForUpdate :one
SELECT *
FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY
UPDATE;

-- name: UpdateAccount :exec
UPDATE accounts
SET balance = $2
WHERE id = $1 RETURNING *;

-- name: AddAccountBalance :exec
UPDATE accounts
SET balance = balance + @balance
WHERE id = @id RETURNING *;

-- name: DeleteAccount :exec
DELETE
FROM accounts
WHERE id = $1;
