package repository

import (
	"context"
	"pr-reviewer-assigment-service/internal/domain"
)

// TeamRepository отвечает за работу с командами.
type TeamRepository interface {
	// Create создаёт новую команду.
	Create(ctx context.Context, team *domain.Team) error

	// GetByName возвращает команду вместе с участниками.
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)
}
