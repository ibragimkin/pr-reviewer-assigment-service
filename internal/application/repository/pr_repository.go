package repository

import (
	"context"
	"pr-reviewer-assigment-service/internal/domain"
)

// PullRequestRepository отвечает за работу с PR.
type PullRequestRepository interface {
	// Create создаёт новый PR.
	Create(ctx context.Context, pr *domain.PullRequest) error

	// GetByID возвращает PR по ID.
	GetByID(ctx context.Context, id string) (*domain.PullRequest, error)

	// Update обновляет существующий PR (например, после merge или reassignment).
	Update(ctx context.Context, pr *domain.PullRequest) error

	// ListByReviewer возвращает список PR'ов, где пользователь назначен ревьювером.
	ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error)
}
