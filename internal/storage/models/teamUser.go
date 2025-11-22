package models

type TeamUser struct {
	internal_id int64  `db:"internal_id"`
	team_id     string `db:"team_id"`
	user_id     string `db:"user_id"`
}
