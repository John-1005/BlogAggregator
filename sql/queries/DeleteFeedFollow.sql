-- name: DeleteFeedFollow :exec
DELETE from feed_follows
WHERE user_id = $1 AND feed_id = $2;
