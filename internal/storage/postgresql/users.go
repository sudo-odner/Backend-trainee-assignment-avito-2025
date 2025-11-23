package postgresql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/domain"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
)

// IsUserReviewerPR Метод проверки, что пользователь является reviewer у PR
func (s *Storage) IsUserReviewerPR(userID, pullRequestID string) error {
	const op = "storage.postgresql.IsUserReviewerPR"

	var exists bool
	err := s.db.QueryRow(
		`select exists(select 1 from pr_reviewers where pull_request_id = $1 and reviewer_id=$2)`,
		userID, pullRequestID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return storage.ErrReviewerNotAssigned
	}
	return nil
}

// GetUserByID Метод получения пользователя
func (s *Storage) GetUserByID(userID string) (*domain.User, error) {
	const op = "storage.postgresql.getUserByID"

	var user domain.User
	row := s.db.QueryRow(`select id, name, is_active from users where id = $1;`, userID)
	if err := row.Scan(&user.ID, &user.Name, &user.IsActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

// GetUserTeamByID Получить команду пользователя
func (s *Storage) GetUserTeamByID(userID string) (string, error) {
	const op = "storage.postgresql.getUserTeamByID"

	var team string
	err := s.db.QueryRow(`select team_name from teams_users where user_id = $1`, userID).Scan(&team)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrTeamNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return team, nil
}

// GetUserPRsByID Получить PRs пользователя
func (s *Storage) GetUserPRsByID(userID string) ([]*domain.PullRequest, error) {
	const op = "storage.postgresql.getUserPRsByID"

	// Получение информации о пользователе
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Получение PRs + информация о reviewers
	rows, err := s.db.Query(
		`
		select pr.id, pr.name, pr.status, pr.merged_at, ru.id, ru.name, ru.is_active
		from pull_requests pr
		left join pr_reviewers r on r.pull_request_id = pr.id
		left join users ru on ru.id = r.reviewer_id
		where author_id = $1
		order by pr.id
		`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	prMap := make(map[string]*domain.PullRequest)
	for rows.Next() {
		var prID, prName, prStatus, reviewerID, reviewerName sql.NullString
		var prMergedAt sql.NullTime
		var reviewerIsActive sql.NullBool

		if err := rows.Scan(
			&prID, &prName, &prStatus, &prMergedAt,
			&reviewerID, &reviewerName, &reviewerIsActive); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		pr, exists := prMap[prID.String]
		if !exists {
			pr = &domain.PullRequest{
				ID:        prID.String,
				Name:      prName.String,
				Author:    *user,
				Status:    prStatus.String,
				Reviewers: []domain.User{},
				MergedAt:  prMergedAt.Time,
			}
			prMap[prID.String] = pr
		}
		if reviewerID.Valid {
			pr.Reviewers = append(pr.Reviewers, domain.User{
				ID:       reviewerID.String,
				Name:     reviewerName.String,
				IsActive: reviewerIsActive.Bool})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userPRs := make([]*domain.PullRequest, 0, len(prMap))
	for _, pr := range prMap {
		userPRs = append(userPRs, pr)
	}
	return userPRs, nil
}

// SetUserIsActive Метод обновления статуса у пользователя
func (s *Storage) SetUserIsActive(userID string, isActive bool) error {
	const op = "storage.postgresql.SetUserIsActive"

	query := `update users set is_active = $1 where id = $2`
	res, err := s.db.Exec(query, isActive, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}

	return nil
}
