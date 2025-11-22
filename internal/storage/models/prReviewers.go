package models

type PrReviewer struct {
	PRID       int64 `db:"pr_id"`
	ReviewerID int64 `db:"reviewer_id"`
}
