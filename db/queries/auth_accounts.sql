-- name: CreateAccount :one
INSERT INTO
    "account" (
        "userId",
        "accountId",
        "providerId",
        password
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreateAccountWithId :one
INSERT INTO
    "account" (
        id,
        "userId",
        "accountId",
        "providerId",
        password
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetAccountByUserIdAndProvider :one
SELECT * FROM "account" WHERE "userId" = $1 AND "providerId" = $2;

-- name: GetAccountById :one
SELECT * FROM "account" WHERE id = $1;

-- name: DeleteAccount :one
DELETE FROM "account" WHERE id = $1 RETURNING *;