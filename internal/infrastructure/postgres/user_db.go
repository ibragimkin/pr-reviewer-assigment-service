package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/domain"
)

type UserDb struct {
	pool *pgxpool.Pool
}

func NewUserDb(pool *pgxpool.Pool) *UserDb {
	return &UserDb{pool: pool}
}

// BulkUpsert создаёт или обновляет нескольких пользователей.
func (r *UserDb) BulkUpsert(ctx context.Context, users []domain.User) (err error) {
	if len(users) == 0 {
		return nil
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	const query = `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			username  = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active
	`

	for _, u := range users {
		if _, err = tx.Exec(ctx, query,
			u.UserID,
			u.Username,
			u.TeamName,
			u.IsActive,
		); err != nil {
			return fmt.Errorf("bulk upsert user %s: %w", u.UserID, err)
		}
	}

	return nil
}

// GetByID возвращает пользователя по user_id.
func (r *UserDb) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	const query = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	var u domain.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&u.UserID,
		&u.Username,
		&u.TeamName,
		&u.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	return &u, nil
}

// SetActive обновляет флаг активности пользователя и возвращает обновлённого пользователя.
func (r *UserDb) SetActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	const query = `
		UPDATE users
		SET is_active = $1
		WHERE user_id = $2
		RETURNING user_id, username, team_name, is_active
	`

	var u domain.User
	err := r.pool.QueryRow(ctx, query, active, userID).Scan(
		&u.UserID,
		&u.Username,
		&u.TeamName,
		&u.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("update user is_active: %w", err)
	}

	return &u, nil
}

// ListByTeam возвращает пользователей команды.
// Если onlyActive == true, возвращаются только активные пользователи.
func (r *UserDb) ListByTeam(ctx context.Context, teamName string, onlyActive bool) ([]domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
	`
	args := []any{teamName}

	if onlyActive {
		query += " AND is_active = TRUE"
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users by team: %w", err)
	}
	defer rows.Close()

	var result []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.UserID,
			&u.Username,
			&u.TeamName,
			&u.IsActive,
		); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}

	return result, nil
}
