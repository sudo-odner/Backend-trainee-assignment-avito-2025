package postgresql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage/models"
)

func initDB(db *sqlx.DB) error {
	schema := `
	create table if not exists users (
	    internal_id bigserial primary key,
	    id text UNIQUE NOT NULL,
	    name text not null,
	    is_active boolean not null
	);
	create table if not exists teams (
	    name text primary key
	);
	create table if not exists teams_users (
	    internal_id bigserial primary key,
	    team_name text references teams(name),
	    user_id text references users(id)
	);
	create table if not exists pull_requests (
	    internal_id bigserial primary key,
	    id text UNIQUE NOT NULL,
	    name text not null,
	    author_id text references users(id),
	    status text not null,
	    merged_at timestamp
	);
	create table if not exists pr_reviewers (
		internal_id bigserial primary key,
	    pull_request_id text references pull_requests(id),
	    reviewer_id text references users(id)
	);`
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

// Добавление пользоватлей, если пользователь с id создан, то обновляем
func (s *Storage) SoftAddUser(id, name string, isActive bool) error {
	const op = "storage.postgresql.SoftAddUser"
	query := `
	insert into users (id, name, is_active) values ($1, $2, $3)
	on conflict (external_id)
	do update set
	   name = $2
	   is_active = $3
	;`
	_, err := s.db.Exec(query, id, name, isActive)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Получить пользователя по id
func (s *Storage) GetUserById(id string) (*models.User, error) {
	const op = "storage.postgresql.GetUserById"
	query := `
	select
	from id, name, is_active
	where id = $1
	;`

	var user models.User
	err := s.db.Get(&user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}
