package postgresql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func initDB(db *sqlx.DB) error {
	schema := `
	create table if not exists users (
	    id bigserial primary key,
	    name text not null,
	    is_active boolean not null
	);
	create table if not exists team (
	    id bigserial primary key,
	    name text not null unique
	);
	create table if not exists pull_requests (
	    id bigserial primary key,
	    name text not null,
	    author_id bigint references users(id),
	    status text not null
	);
	create table if not exists pr_reviewers (
	    pull_request_id bigint references pull_requests(id),
	    reviewer_id bigint references users(id)
	);
`
	_, err := db.Exec(schema)
	return err
}

type Storage struct {
	db *sqlx.DB
}

func New(host, port, user, password, dbName, sslMode string) (*Storage, error) {
	const op = "storage.postgresql.New"

	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)
	db, err := sqlx.Open("postgres", dns)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = initDB(db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}
