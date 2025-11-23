package storage

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

func RunMigrations(db *sqlx.DB) error {
	data, err := os.ReadFile("./migrations/001_init.up.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}

	_, err = db.Exec(string(data))
	if err != nil {
		return fmt.Errorf("exec migration: %w", err)
	}

	return nil
}
