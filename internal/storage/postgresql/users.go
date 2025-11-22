package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/domain"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
)

func userIsCreated(tx *sql.Tx, userID string) (*domain.User, error) {
	const op = "storage.userIsCreated"

	var user domain.User
	row := tx.QueryRow(`select id, name, is_active from users where id = $1;`, userID)
	if err := row.Scan(&user.ID, &user.Name, &user.IsActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

func (s *Storage) SetUserIsActive(userID string, isActive bool) error {
	const op = "storage.postgresql.SetUserIsActive"

	query := `UPDATE users SET is_active = $1 WHERE id = $2`
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

func (s *Storage) GetUserPRsByUserID(userID string) ([]domain.PullRequest, error) {
	const op = "storage.postgresql.GetUserPRsByUserID"
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	querySelectPRs := `
	select pr.id, pr.name, pr.status, pr.merged_at, ru.id, ru.name, ru.is_active
    from pull_requests pr 
	left join pr_reviewers r on pr.id = r.pull_request_id
    left join users ru on ru.id = r.reviewer_id
	where pr.author_id = $1
	order by pr.id
	;`

	// Поиск и проверка пользователя
	user, err := userIsCreated(tx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Поиск всех связанных к нему PR
	rows, err := tx.Query(querySelectPRs, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	prMap := make(map[string]*domain.PullRequest)
	for rows.Next() {
		var prID, prName, prStatus, reviewerID, reviewerName sql.NullString
		var mergedAt sql.NullTime
		var reviewedIsActive sql.NullBool
		if err := rows.Scan(
			&prID,
			&prName,
			&prStatus,
			&mergedAt,
			&reviewerID,
			&reviewerName,
			&reviewedIsActive); err != nil {
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
				MergedAt:  mergedAt.Time,
			}
			prMap[prID.String] = pr
		}
		if reviewerID.Valid {
			pr.Reviewers = append(pr.Reviewers, domain.User{
				ID:       reviewerID.String,
				Name:     reviewerName.String,
				IsActive: reviewedIsActive.Bool})
		}
	}

	userPRs := make([]domain.PullRequest, 0, len(prMap))
	for _, pr := range prMap {
		userPRs = append(userPRs, *pr)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return userPRs, nil
}

func (s *Storage) GetUserByID(userID string) (*domain.User, error) {
	const op = "storage.postgresql.GetUserByID"
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()
	user, err := userIsCreated(tx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}
