-- name: ListFeeds :many
SELECT feeds.name as feed_name, feeds.url, users.name
FROM feeds
JOIN users ON feeds.user_id = users.id;
