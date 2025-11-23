package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/domain"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
)

// IsTeamExists Проверка существования команды
func (s *Storage) IsTeamExists(nameTeam string) error {
	const op = "storage.postgresql.IsTeamExists"

	var exists bool
	err := s.db.QueryRow(`select exists(select 1 from teams where name = $1)`, nameTeam).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return fmt.Errorf("%s: %w", op, storage.ErrTeamNotFound)
	}
	return nil

}

// GetTeam Получение команды и ее пользователей
func (s *Storage) GetTeam(nameTeam string) (*domain.Team, error) {
	const op = "storage.postgresql.GetTeam"

	if err := s.IsTeamExists(nameTeam); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Получение команды и ее пользователей
	rows, err := s.db.Query(`
		select u.id, u.name, u.is_active
			from teams_users tu
			left join users u on tu.user_id = u.id
			where tu.team_name = $1;`,
		nameTeam)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var team domain.Team

	team.Name = nameTeam
	team.Users = make([]domain.User, 0)
	for rows.Next() {
		var userID, userName sql.NullString
		var isActive sql.NullBool
		if err = rows.Scan(&userID, &userName, &isActive); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		team.Users = append(team.Users, domain.User{ID: userID.String, Name: userName.String, IsActive: isActive.Bool})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &team, nil
}

// CreateTeamWithUser Создание команды и добавление пользователь в нее
func (s *Storage) CreateTeamWithUser(nameTeam string, users []domain.User) error {
	const op = "storage.postgresql.CreateTeamWithUser"
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
	querySoftTeamsUsers := `insert into teams_users (team_name, user_id) values ($1, $2) on conflict do nothing`

	// Создаем команду
	res := tx.QueryRow(queryInsertTeam, nameTeam)
	var nameTeamRes string
	if err = res.Scan(&nameTeamRes); err != nil {
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
		_, err = tx.Exec(querySoftTeamsUsers, nameTeam, user.ID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
