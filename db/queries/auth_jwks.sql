-- name: GetJwksSets :many
SELECT
    id,
    "publicKey",
    "createdAt",
    "expiresAt"
FROM jwks
WHERE
    "expiresAt" IS NULL
    OR "expiresAt" > NOW();