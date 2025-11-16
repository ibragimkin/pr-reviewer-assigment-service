package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/domain"
)

type TeamDb struct {
	pool *pgxpool.Pool
}

func NewTeamDb(pool *pgxpool.Pool) *TeamDb {
	return &TeamDb{pool: pool}
}

// Create создаёт новую команду.
// Если команда с таким именем уже существует - возвращает repository.ErrAlreadyExists.
func (r *TeamDb) Create(ctx context.Context, team *domain.Team) error {
	const query = `
		INSERT INTO teams (team_name)
		VALUES ($1)
	`

	_, err := r.pool.Exec(ctx, query, team.TeamName)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return repository.ErrAlreadyExists
		}
		return fmt.Errorf("insert team %s: %w", team.TeamName, err)
	}

	return nil
}

// GetByName возвращает команду вместе с участниками.
// Если команда не найдена - возвращает repository.ErrNotFound.
func (r *TeamDb) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	const queryTeam = `
		SELECT team_name
		FROM teams
		WHERE team_name = $1
	`

	var name string
	err := r.pool.QueryRow(ctx, queryTeam, teamName).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("query team by name: %w", err)
	}

	const queryMembers = `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`

	rows, err := r.pool.Query(ctx, queryMembers, teamName)
	if err != nil {
		return nil, fmt.Errorf("query team members: %w", err)
	}
	defer rows.Close()

	members := make([]domain.TeamMember, 0)
	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team members: %w", err)
	}

	team := &domain.Team{
		TeamName: name,
		Members:  members,
	}

	return team, nil
}
