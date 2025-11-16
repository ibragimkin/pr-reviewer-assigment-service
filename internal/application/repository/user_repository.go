package repository

import (
	"context"
	"pr-reviewer-assigment-service/internal/domain"
)

// UserRepository отвечает за работу с пользователями.
type UserRepository interface {
	// BulkUpsert создаёт или обновляет нескольких пользователей.
	BulkUpsert(ctx context.Context, users []domain.User) error

	// GetByID возвращает пользователя по user_id.
	GetByID(ctx context.Context, userID string) (*domain.User, error)

	// SetActive обновляет флаг активности пользователя и возвращает обновлённого пользователя.
	SetActive(ctx context.Context, userID string, active bool) (*domain.User, error)

	// ListByTeam возвращает пользователей команды.
	ListByTeam(ctx context.Context, teamName string, onlyActive bool) ([]domain.User, error)
}
