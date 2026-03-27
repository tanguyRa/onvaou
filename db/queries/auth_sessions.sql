-- CreateSession(ctx context.Context, session *Session) error
-- GetSession(ctx context.Context, token string) (*Session, error)
-- UpdateSession(ctx context.Context, token string, updates map[string]any) error
-- DeleteSession(ctx context.Context, token string) error
-- DeleteUserSessions(ctx context.Context, userID string) error

-- Sessions table
-- CREATE TABLE "session" (
--     id UUID PRIMARY KEY,
--     token VARCHAR(255) UNIQUE NOT NULL,
--     user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
--     expires_at TIMESTAMPTZ NOT NULL,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     ip_address VARCHAR(45) NOT NULL,
--     user_agent VARCHAR NOT NULL,
--     impersonated_by UUID REFERENCES users (id)
-- );

-- name: CreateSession :one
INSERT INTO
    "session" (
        token,
        "userId",
        "expiresAt",
        "ipAddress",
        "userAgent"
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: CreateSessionWithId :one
INSERT INTO
    "session" (
        id,
        token,
        "userId",
        "expiresAt",
        "ipAddress",
        "userAgent"
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: GetSession :one
SELECT * FROM "session" WHERE id = $1;

-- name: UpdateSession :one
UPDATE "session"
SET
    token = $2,
    "userId" = $3,
    "expiresAt" = $4,
    "ipAddress" = $5,
    "userAgent" = $6,
    "updatedAt" = CURRENT_TIMESTAMP
WHERE
    id = $1
RETURNING
    *;

-- name: DeleteSession :one
DELETE FROM "session" WHERE id = $1 RETURNING *;

-- name: DeleteUserSessions :exec
DELETE FROM "session" WHERE "userId" = $1;

-- name: GetSessionByToken :one
SELECT * FROM "session" WHERE token = $1;