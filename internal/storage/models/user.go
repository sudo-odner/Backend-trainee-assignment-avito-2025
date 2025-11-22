package models

type User struct {
	InternalID int64  `db:"internal_id"`
	ID         string `db:"id"`
	Name       string `db:"name"`
	IsActive   bool   `db:"is_active"`
}
