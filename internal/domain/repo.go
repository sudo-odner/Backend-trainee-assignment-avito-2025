package domain

type Storage interface {
	CreateTeam(team Team, users []User) error
	GetUsersTeamByNameTeam(teamName string) ([]User, error)
}
