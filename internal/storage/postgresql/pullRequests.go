package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/domain"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
)

// IsMergePR Проверка, что мердж существует
func (s *Storage) IsMergePR(prID string) error {
	const op = "storage.postgresql.IsMergePR"
	// Проверяем что PR, еще не merged
	var status string
	err := s.db.QueryRow(`select status from pull_requests where id=$1`, prID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrPRNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	if status == "MERGED" {
		return storage.ErrPRAlreadyMerged
	}

	return nil
}

// MergePR Создание мердж для pr
func (s *Storage) MergePR(prID string) error {
	const op = "storage.postgresql.MergePR"

	// Проверяем что PR, еще не merged
	if err := s.IsMergePR(prID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	// Обновляем merge
	_, err := s.db.Exec(`update pull_requests set status = $2, merged_at = $3 where id = $1`, prID, "MERGED", time.Now())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetPRByID Получение PR по id
func (s *Storage) GetPRByID(pullRequestID string) (*domain.PullRequest, error) {
	const op = "storage.postgresql.GetPRByID"

	querySelectPR := `
	select pr.id, pr.name, a.id, a.name, a.is_active, pr.status, pr.merged_at, ru.id, ru.name, ru.is_active
	from pull_requests pr
	left join users a on a.id = pr.author_id
	left join pr_reviewers r on pr.id = r.pull_request_id
	left join users ru on r.reviewer_id = ru.id
	where pr.id = $1
	order by pr.id
	`

	rows, err := s.db.Query(querySelectPR, pullRequestID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var pr *domain.PullRequest

	reviewersMap := make(map[string]domain.User)
	if rows.Next() {
		var prID, prName, authorID, authorName, prStatus, reviewerID, reviewerName sql.NullString
		var authorIsActive, reviewerIsActive sql.NullBool
		var prMergedAt sql.NullTime
		if err := rows.Scan(
			&prID,
			&prName,
			&authorID, &authorName, &authorIsActive,
			&prStatus, &prMergedAt,
			&reviewerID, &reviewerName, reviewerIsActive); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if pr == nil {
			pr = &domain.PullRequest{
				ID:   prID.String,
				Name: prName.String,
				Author: domain.User{
					ID:       authorID.String,
					Name:     authorName.String,
					IsActive: authorIsActive.Bool,
				},
				Status:    prStatus.String,
				Reviewers: []domain.User{},
				MergedAt:  prMergedAt.Time,
			}
		}

		if reviewerID.Valid {
			if _, exists := reviewersMap[reviewerID.String]; !exists {
				reviewer := domain.User{
					ID:       reviewerID.String,
					Name:     reviewerName.String,
					IsActive: reviewerIsActive.Bool,
				}
				pr.Reviewers = append(pr.Reviewers, reviewer)
				reviewersMap[reviewerID.String] = reviewer
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if pr == nil {
		return nil, storage.ErrPRNotFound
	}

	return pr, nil
}

// CreatePRWithReviewers Создание PR c автоматически рандомно назначеными reviewer
func (s *Storage) CreatePRWithReviewers(prID, prName, authorID string) (*domain.PullRequest, error) {
	const op = "storage.postgresql.CreatePRWithReviewers"
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Получаем автора (проверка на существования)
	author, err := s.GetUserByID(authorID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создаем пулреквест, если уже создан то отменяем все
	_, err = tx.Exec(
		`insert into pull_requests (id, name, author_id, status) values ($1, $2, $3, $4)`,
		prID, prName, authorID, "OPEN")
	if err != nil {
		_ = tx.Rollback()
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, storage.ErrPRAlreadyExists
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Получаем команду пользователя
	nameTeam, err := s.GetUserTeamByID(author.ID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Выбираем до 2 активных ревюверов из команды автора, кроме него
	rows, err := tx.Query(`
	select u.id, u.name, u.is_active
	from teams_users tu
	join users u on tu.user_id = u.id
	where tu.team_name = $1 and u.is_active=true and u.id <> $2
	order by random() limit 2`,
		nameTeam, authorID,
	)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var reviewers []domain.User
	for rows.Next() {
		var r domain.User
		if err := rows.Scan(&r.ID, &r.Name, &r.IsActive); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		reviewers = append(reviewers, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создаем связи
	for _, r := range reviewers {
		if _, err := tx.Exec(`insert into pr_reviewers(pull_request_id, reviewer_id) values($1, $2)`, prID, r.ID); err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &domain.PullRequest{
		ID:        prID,
		Name:      prName,
		Author:    *author,
		Status:    "OPEN",
		Reviewers: reviewers,
	}, nil
}

// ReassignReviewer Переназначение reviewer, если это возможно
func (s *Storage) ReassignReviewer(prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	const op = "storage.postgresql.ReassignReviewer"

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Получаем пользователя (проверка на его существования)
	_, err = s.GetUserByID(oldReviewerID)
	if err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на MERGE PR
	if err := s.IsMergePR(prID); err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Проверка на то что пользователь назначен как reviewer
	if err := s.IsUserReviewerPR(prID, oldReviewerID); err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Получаем название команды у reviewer
	teamName, err := s.GetUserTeamByID(oldReviewerID)
	if err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Находим случайного активного пользователя из команды (кроме старого reviewer)
	var newReviewerID string
	err = tx.QueryRow(`
	select u.id 
    from teams_users tu
    join users u ON tu.user_id = u.id
    where tu.team_name=$1 and u.is_active=true and u.id <> $2 and u.id <> (select author_id from pull_requests where id = $3)
    order by random() limit 1`, teamName, oldReviewerID, prID).Scan(&newReviewerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = tx.Rollback()
			return nil, "", storage.ErrNoCandidate
		}
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Обновляем reviewer
	_, err = tx.Exec(
		`update pr_reviewers set reviewer_id=$1 where pull_request_id=$2 AND reviewer_id=$3`,
		newReviewerID, prID, oldReviewerID)
	if err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	// Получаем обновлённый PR с reviewer
	pr, err := s.GetPRByID(prID)
	if err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}
	return pr, newReviewerID, nil
}
