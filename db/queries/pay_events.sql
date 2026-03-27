-- name: CreateEvent :one
INSERT INTO
    "system_events" ("userId", "type", "data")
VALUES ($1, $2, $3)
RETURNING
    *;

-- name: CreateEventWithId :one
INSERT INTO
    "system_events" (
        "id",
        "userId",
        "type",
        "data"
    )
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetEventByID :one
SELECT * FROM "system_events" WHERE id = $1;

-- name: GetEventByUserIDAndType :one
SELECT * FROM "system_events" WHERE "userId" = $1 AND "type" = $2;

-- name: ListEventsByUserID :many
SELECT * FROM "system_events" WHERE "userId" = $1 ORDER BY "createdAt" DESC;
