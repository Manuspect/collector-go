-- name: IsCreatedRefreshTokenDb :one
SELECT token
FROM tokens
WHERE user_id = $1;
-- name: SaveRefreshToken :one
INSERT INTO tokens (user_id, token)
VALUES($1, $2)
RETURNING *;
-- name: UpdateRefreshTokenDb :exec
UPDATE tokens
SET token = $1
WHERE user_id = $2
RETURNING *;
-- name: GetRefreshTokensDb :many
SELECT id,
    user_id,
    token
FROM tokens;
-- name: DeleteRefreshTokenDb :exec
DELETE FROM tokens
WHERE id = $1;