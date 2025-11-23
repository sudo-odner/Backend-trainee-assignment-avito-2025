package postgresql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
)

type Storage struct {
	db *sqlx.DB
}

func New(host, port, user, password, dbName, sslMode string) (*Storage, error) {
	const op = "storage.postgresql.New"

	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)
	// Открываем соединение
	db, err := sqlx.Open("postgres", dns)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// Миграция
	if err := storage.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}
