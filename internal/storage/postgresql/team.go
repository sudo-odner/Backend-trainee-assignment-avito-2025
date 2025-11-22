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
	err := tx.QueryRow(`select team_name from teams_users where user_id = $1`, userID).Scan(&team)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrTeamNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return team, nil
}

func teamExists(tx *sql.Tx, nameTeam string) error {
	const op = "storage.teamExists"

	var exists bool
	err := tx.QueryRow(`select exists(select 1 from teams where name = $1)`, nameTeam).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return fmt.Errorf("%s: %w", op, storage.ErrTeamNotFound)
	}
	return nil
}

// Создание команды и добавление пользователь в нее
func (s *Storage) CreateTeam(nameTeam string, users []domain.User) error {
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

// Получение команды пользователя
func (s *Storage) GetTeamByUserID(userID string) (string, error) {
	const op = "storage.getTeamByUserID"
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	team, err := getUserTeam(tx, userID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return team, nil
}

// Получение всех пользователей у команды по ее имени
func (s *Storage) GetUsersTeamByName(nameTeam string) (*domain.Team, error) {
	const op = "storage.postgresql.GetUsersTeamByName"
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	if err := teamExists(tx, nameTeam); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	query := `
	select u.internal_id, u.id, u.name, u.is_active
	from teams_users tu
	left join users u on tu.user_id = u.id
	where tu.team_name = $1
	;`

	rows, err := tx.Query(query, nameTeam)
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

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &team, nil
}
