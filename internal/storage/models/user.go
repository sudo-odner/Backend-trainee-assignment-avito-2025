package models

type User struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	isActive bool   `db:"is_active"`
}
