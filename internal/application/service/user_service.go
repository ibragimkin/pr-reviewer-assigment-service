package service

import (
	"context"
	"errors"
	"fmt"
	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/domain"
)

type UserService struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
}

func NewUserService(
	userRepository repository.UserRepository,
	prRepository repository.PullRequestRepository,
) *UserService {
	return &UserService{
		userRepo: userRepository,
		prRepo:   prRepository,
	}
}

// SetIsActive устанавливает флаг активности пользователя и возвращает обновлённого пользователя.
func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.SetActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(domain.ErrorNotFound, "user not found: "+userID)
		}
		return nil, fmt.Errorf("userRepo.SetActive: %w", err)
	}

	return user, nil
}

// GetReview возвращает список PR'ов, где пользователь назначен ревьювером.
func (s *UserService) GetReview(ctx context.Context, userID string) (string, []domain.PullRequestShort, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil, domain.NewError(domain.ErrorNotFound, "user not found: "+userID)
		}
		return "", nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}
	prs, err := s.prRepo.ListByReviewer(ctx, userID)
	if err != nil {
		return "", nil, fmt.Errorf("prRepo.ListByReviewer: %w", err)
	}
	return userID, prs, nil
}
