-- name: CreateEvent :one
INSERT INTO
    "events" ("userId", "type", "data")
VALUES ($1, $2, $3)
RETURNING
    *;

-- name: CreateEventWithId :one
INSERT INTO
    "events" (
        "id",
        "userId",
        "type",
        "data"
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetEventByID :one
SELECT * FROM "events" WHERE id = $1;

-- name: GetEventByUserIDAndType :one
SELECT * FROM "events" WHERE "userId" = $1 AND "type" = $2;

-- name: ListEventsByUserID :many
SELECT * FROM "events" WHERE "userId" = $1 ORDER BY "createdAt" DESC;