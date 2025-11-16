package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/domain"
)

type PullRequestDb struct {
	pool *pgxpool.Pool
}

func NewPullRequestDb(pool *pgxpool.Pool) *PullRequestDb {
	return &PullRequestDb{pool: pool}
}

// Create создаёт новый PR.
// Если PR с таким ID уже существует - возвращает repository.ErrAlreadyExists.
func (r *PullRequestDb) Create(ctx context.Context, pr *domain.PullRequest) error {
	const query = `
		INSERT INTO pull_requests (
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			assigned_reviewers,
			created_at,
			merged_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	createdAt := time.Now().UTC()
	if pr.CreatedAt != nil {
		createdAt = *pr.CreatedAt
	}

	var mergedAt any
	if pr.MergedAt != nil {
		mergedAt = *pr.MergedAt
	} else {
		mergedAt = nil
	}

	_, err := r.pool.Exec(ctx, query,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
		pr.AssignedReviewers,
		createdAt,
		mergedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrAlreadyExists
		}
		return fmt.Errorf("insert pull_request %s: %w", pr.PullRequestID, err)
	}

	return nil
}

// GetByID возвращает PR по ID.
// Если не найден - возвращает repository.ErrNotFound.
func (r *PullRequestDb) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	const query = `
		SELECT
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			assigned_reviewers,
			created_at,
			merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	var (
		pr        domain.PullRequest
		createdAt pgtype.Timestamptz
		mergedAt  pgtype.Timestamptz
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.AssignedReviewers,
		&createdAt,
		&mergedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("query pull_request by id: %w", err)
	}

	if createdAt.Valid {
		t := createdAt.Time
		pr.CreatedAt = &t
	}
	if mergedAt.Valid {
		t := mergedAt.Time
		pr.MergedAt = &t
	}

	return &pr, nil
}

// Update обновляет существующий PR (например, после merge или reassignment).
// Если PR не найден - возвращает repository.ErrNotFound.
func (r *PullRequestDb) Update(ctx context.Context, pr *domain.PullRequest) error {
	const query = `
		UPDATE pull_requests
		SET
			pull_request_name   = $2,
			author_id           = $3,
			status              = $4,
			assigned_reviewers  = $5,
			created_at          = $6,
			merged_at           = $7
		WHERE pull_request_id = $1
	`

	createdAt := time.Now().UTC()
	if pr.CreatedAt != nil {
		createdAt = *pr.CreatedAt
	}

	var mergedAt any
	if pr.MergedAt != nil {
		mergedAt = *pr.MergedAt
	} else {
		mergedAt = nil
	}

	cmdTag, err := r.pool.Exec(ctx, query,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
		pr.AssignedReviewers,
		createdAt,
		mergedAt,
	)
	if err != nil {
		return fmt.Errorf("update pull_request %s: %w", pr.PullRequestID, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// ListByReviewer возвращает список PR'ов, где пользователь назначен ревьювером.
func (r *PullRequestDb) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	const query = `
		SELECT
			pull_request_id,
			pull_request_name,
			author_id,
			status
		FROM pull_requests
		WHERE $1 = ANY(assigned_reviewers)
		ORDER BY pull_request_id
	`

	rows, err := r.pool.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("list pull_requests by reviewer: %w", err)
	}
	defer rows.Close()

	var result []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
		); err != nil {
			return nil, fmt.Errorf("scan pull_request row: %w", err)
		}
		result = append(result, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pull_request rows: %w", err)
	}

	return result, nil
}

// GetReviewerStats получает статистику назначений по ревьюверам.
func (r *PullRequestDb) GetReviewerStats(ctx context.Context) ([]domain.ReviewerStat, error) {
	const query = `
        SELECT reviewer_id, COUNT(*) AS review_count
        FROM (
            SELECT unnest(assigned_reviewers) AS reviewer_id
            FROM pull_requests
        ) AS t
        GROUP BY reviewer_id
        ORDER BY reviewer_id
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query reviewer stats: %w", err)
	}
	defer rows.Close()

	stats := make([]domain.ReviewerStat, 0)

	for rows.Next() {
		var stat domain.ReviewerStat
		if err := rows.Scan(&stat.UserID, &stat.ReviewCount); err != nil {
			return nil, fmt.Errorf("scan reviewer stat: %w", err)
		}
		stats = append(stats, stat)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows error: %w", rows.Err())
	}

	return stats, nil
}
