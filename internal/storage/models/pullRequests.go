package models

type pullRequests struct {
	ID        int64   `db:"id"`
	Name      string  `db:"name"`
	AutorID   int64   `db:"autor_id"`
	Status    string  `db:"status"`
	Reviewers []int64 `db:"reviewers"`
}
