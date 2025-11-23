package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/domain"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
)

// DeactivateTeamUsers Массовая деактивация пользователей команды
func (s *Storage) DeactivateTeamUsers(teamName string) (int, error) {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return -1, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("tx rollback failed: %v", err)
		}
	}()

	// Проверка на существоание команды
	if err := s.IsTeamExists(teamName); err != nil {
		return -1, err
	}

	// Деактивируем пользователей
	res, err := tx.Exec(`
        update users u
        set is_active = false
        from teams_users tu
        where tu.team_name = $1 and u.id = tu.user_id
    `, teamName)
	if err != nil {
		return -1, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}

	// Удаляем их из pr_reviewers для открытых PR
	_, err = tx.Exec(`
		delete from pr_reviewers
		using pull_requests, teams_users
		where pr_reviewers.reviewer_id = teams_users.user_id
		and teams_users.team_name = $1
		and pr_reviewers.pull_request_id = pull_requests.id
		and pull_requests.status = 'OPEN';
    `, teamName)
	if err != nil {
		return -1, err
	}

	if err := tx.Commit(); err != nil {
		return -1, err
	}

	return int(rowsAffected), nil
}

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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows close failed: %v", err)
		}
	}()

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
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("tx rollback failed: %v", err)
		}
	}()

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
		return fmt.Errorf("%s: %w", op, err)
	}

	// Добавляем пользователей(обновляем/создаем пользователя) к команде
	for _, user := range users {
		// Обновляем/Добовляем пользователей
		_, err = tx.Exec(querySoftInsertUser, user.ID, user.Name, user.IsActive)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		// Удаляем старую связь с командой (у пользователя должна быть одна команда) TODO: Придумать другую логику
		if _, err := tx.Exec(`delete from teams_users where user_id = $1`, user.ID); err != nil {
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
