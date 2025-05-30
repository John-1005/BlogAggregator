// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: names.sql

package database

import (
	"context"
)

const getName = `-- name: GetName :one


SELECT id, created_at, updated_at, name FROM USERS WHERE name = $1
`

func (q *Queries) GetName(ctx context.Context, name string) (User, error) {
	row := q.db.QueryRowContext(ctx, getName, name)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
	)
	return i, err
}
