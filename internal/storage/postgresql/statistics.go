package postgresql

import (
	"fmt"
	"log"

	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/domain"
)

func (s *Storage) GetReviewStat() ([]domain.UserReviewStat, error) {
	const op = "storage.GetReviewsStat"
	rows, err := s.db.Query(`
        select u.id as user_id, count(r.reviewer_id) AS review_count
        from users u
        left join pr_reviewers r on u.id = r.reviewer_id
        group by u.id
        order by review_count desc;
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows close failed: %v", err)
		}
	}()

	// Переброр всех reviewer
	stats := make([]domain.UserReviewStat, 0)
	for rows.Next() {
		var stat domain.UserReviewStat
		err := rows.Scan(&stat.UserID, &stat.ReviewCount)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return stats, nil
}
