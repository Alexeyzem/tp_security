package repository

import (
	"context"
	"database/sql"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

const querySave = "INSERT INTO table (request, response) VALUES ($1, $2)"

func (r *Repo) Save(ctx context.Context, request, response string) error {
	_, err := r.db.ExecContext(ctx, querySave, request, response)

	return err
}
