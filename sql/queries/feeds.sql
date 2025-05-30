


-- name: CreateFeed :one
INSERT INTO feeds(id, user_id, name, created_at, updated_at, url)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
Returning *;
