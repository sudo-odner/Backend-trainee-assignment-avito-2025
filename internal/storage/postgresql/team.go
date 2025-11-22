package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/domain"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage/models"
)

func getUserTeam(tx *sql.Tx, userID string) (string, error) {
	const op = "storage.getUserTeam"
	var team string
	err := tx.QueryRow(`select team_name from teams where user_id = &1`, userID).Scan(&team)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrUserNotFound
		}
		return "", err
	}
	return team, nil
}

// Создание команды и добавление пользователь в нее
func (s *Storage) CreateTeam(team domain.Team, users []domain.User) error {
	const op = "storage.postgresql.CreateTeam"
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	queryInsertTeam := `insert into teams (name) values ($1) on conflict(name) do nothing returning name`
	querySoftInsertUser := `
	insert into users (id, name, is_active) values ($1, $2, $3)
	on conflict (id) do update set 
	    name = EXCLUDED.name,
	    is_active = EXCLUDED.is_active
	`
	querySoftTeamsUsers := `insert into users (id, name, is_active) values ($1, $2, $3) on conflict do nothing`

	// Создаем команду
	res := tx.QueryRow(queryInsertTeam, team.Name)
	var nameTeam string
	if err = res.Scan(&nameTeam); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrTeamAlreadyExists
		}
	}

	// Добавляем пользователей(обновляем/создаем пользователя) к команде
	for _, user := range users {
		// Обновляем/Добовляем пользователей
		_, err = tx.Exec(querySoftInsertUser, user.ID, user.Name, user.IsActive)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		// Добавляем связи
		_, err = tx.Exec(querySoftTeamsUsers, team.Name, user.ID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Получение всех пользователей у команды по ее имени
func (s *Storage) GetUsersTeamByName(nameTeam string) (*domain.Team, error) {
	const op = "storage.postgresql.GetUsersTeamByName"

	query := `
	select u.internal_id, u.id, u.name, u.is_active
	from teams_users tu
	left join users u on tu.user_id = u.id
	where tu.team_name = $1
	;`

	rows, err := s.db.Query(query, nameTeam)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var team domain.Team
	team.Name = nameTeam
	team.Users = make([]domain.User, 0)
	for rows.Next() {
		var u models.User
		if err = rows.Scan(&u.InternalID, &u.ID, &u.Name, &u.IsActive); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		team.Users = append(team.Users, domain.User{ID: u.ID, Name: u.Name, IsActive: u.IsActive})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &team, nil
}
