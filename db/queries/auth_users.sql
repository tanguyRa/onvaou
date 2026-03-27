-- name: ListUsers :many
SELECT * FROM "user";

-- name: CreateUser :one
INSERT INTO
    "user" (
        "name",
        "email",
        "emailVerified",
        "image"
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: CreateUserWithId :one
INSERT INTO
    "user" (
        "id",
        "name",
        "email",
        "emailVerified",
        "image"
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetUserByID :one
SELECT * FROM "user" WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = $1;

-- name: UpdateUser :one
UPDATE "user"
SET
    "name" = $2,
    "email" = $3,
    "emailVerified" = $4,
    "image" = $5,
    "updatedAt" = CURRENT_TIMESTAMP
WHERE
    id = $1
RETURNING
    *;

-- name: DeleteUser :one
DELETE FROM "user" WHERE id = $1 RETURNING *;