package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net/url"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(postgresURL, user, password, dbname string) (*Repository, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		url.QueryEscape(user),
		url.QueryEscape(password),
		postgresURL,
		dbname,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}
