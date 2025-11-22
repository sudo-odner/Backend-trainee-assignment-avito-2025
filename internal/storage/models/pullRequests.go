package models

type pullRequests struct {
	InternalId int64    `db:"internal_id"`
	ID         string   `db:"id"`
	Name       string   `db:"name"`
	AutorID    string   `db:"autor_id"`
	Status     string   `db:"status"`
	Reviewers  []string `db:"reviewers"`
}
