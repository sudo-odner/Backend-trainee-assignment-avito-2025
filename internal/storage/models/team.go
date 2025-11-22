package models

type Team struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}
