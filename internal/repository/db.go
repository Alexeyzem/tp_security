package repository

import (
	"context"
	"database/sql"
	"errors"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

const (
	querySave    = "INSERT INTO data (request, response) VALUES ($1, $2)"
	queryGetByID = "SELECT request, response FROM data WHERE id = $1"
	queryGet     = "SELECT id, request, response FROM data"
)

func (r *Repo) Save(ctx context.Context, request, response string) error {
	_, err := r.db.ExecContext(ctx, querySave, request, response)

	return err
}

func (r *Repo) GetOne(ctx context.Context, id int) (req string, resp string, err error) {
	row := r.db.QueryRowContext(ctx, queryGetByID, id)
	err = row.Scan(&req, &resp)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}

	if err != nil {
		return "", "", err
	}

	return req, resp, nil
}

func (r *Repo) GetAll(ctx context.Context) (map[int][2]string, error) {
	rows, err := r.db.QueryContext(ctx, queryGet)
	if err != nil {
		return nil, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	defer rows.Close()
	res := map[int][2]string{}

	for rows.Next() {
		var id int
		var req string
		var resp string
		if err := rows.Scan(&id, &req, &resp); err != nil {
			return nil, err
		}
		res[id] = [2]string{req, resp}
	}

	return res, nil
}
