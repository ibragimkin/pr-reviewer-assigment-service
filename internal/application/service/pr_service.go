package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/domain"
)

type PullRequestService struct {
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
}

func NewPullRequestService(
	prRepository repository.PullRequestRepository,
	userRepository repository.UserRepository,
	teamRepository repository.TeamRepository,
) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepository,
		userRepo: userRepository,
		teamRepo: teamRepository,
	}
}

// Create создаёт PR и назначает до двух активных ревьюверов из команды автора (исключая самого автора).
func (s *PullRequestService) Create(
	ctx context.Context,
	prID string,
	prName string,
	authorID string,
) (*domain.PullRequest, error) {
	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(domain.ErrorNotFound, "author not found: "+authorID)
		}
		return nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}

	team, err := s.teamRepo.GetByName(ctx, author.TeamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(domain.ErrorNotFound, "team not found: "+author.TeamName)
		}
		return nil, fmt.Errorf("teamRepo.GetByName: %w", err)
	}

	_ = team

	users, err := s.userRepo.ListByTeam(ctx, author.TeamName, true)
	if err != nil {
		return nil, fmt.Errorf("userRepo.ListByTeam: %w", err)
	}

	reviewerIDs := make([]string, 0, 2)
	for _, u := range users {
		if u.UserID == authorID {
			continue
		}
		reviewerIDs = append(reviewerIDs, u.UserID)
		if len(reviewerIDs) == 2 {
			break
		}
	}

	now := time.Now().UTC()

	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: reviewerIDs,
		CreatedAt:         &now,
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, domain.NewError(domain.ErrorPRExists, "pull request already exists: "+prID)
		}
		return nil, fmt.Errorf("prRepo.Create: %w", err)
	}

	return pr, nil
}

// Merge помечает PR как MERGED (идемпотентно).
func (s *PullRequestService) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(domain.ErrorNotFound, "pull request not found: "+prID)
		}
		return nil, fmt.Errorf("prRepo.GetByID: %w", err)
	}

	if pr.Status == "MERGED" {
		if pr.MergedAt == nil {
			now := time.Now().UTC()
			pr.MergedAt = &now
			if err := s.prRepo.Update(ctx, pr); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return nil, domain.NewError(domain.ErrorNotFound, "pull request not found on update: "+prID)
				}
				return nil, fmt.Errorf("prRepo.Update: %w", err)
			}
		}
		return pr, nil
	}

	now := time.Now().UTC()
	pr.Status = "MERGED"
	pr.MergedAt = &now

	if err := s.prRepo.Update(ctx, pr); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(domain.ErrorNotFound, "pull request not found on update: "+prID)
		}
		return nil, fmt.Errorf("prRepo.Update: %w", err)
	}
	return pr, nil
}

// Reassign переносит одного ревьювера на другого из его команды.
// После MERGED менять ревьюверов нельзя.
// Если ревьювер не назначен - NOT_ASSIGNED.
// Если нет доступных кандидатов - NO_CANDIDATE.
func (s *PullRequestService) Reassign(
	ctx context.Context,
	prID string,
	oldUserID string,
) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", domain.NewError(domain.ErrorNotFound, "pull request not found: "+prID)
		}
		return nil, "", fmt.Errorf("prRepo.GetByID: %w", err)
	}

	if pr.Status == "MERGED" {
		return nil, "", domain.NewError(domain.ErrorPRMerged, "cannot reassign reviewers on merged PR")
	}
	found := false
	for _, r := range pr.AssignedReviewers {
		if r == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return nil, "", domain.NewError(domain.ErrorNotAssigned, "user is not assigned as reviewer on this PR")
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldUserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", domain.NewError(domain.ErrorNotFound, "reviewer not found: "+oldUserID)
		}
		return nil, "", fmt.Errorf("userRepo.GetByID: %w", err)
	}

	users, err := s.userRepo.ListByTeam(ctx, oldReviewer.TeamName, true)
	if err != nil {
		return nil, "", fmt.Errorf("userRepo.ListByTeam: %w", err)
	}

	candidates := make([]domain.User, 0, len(users))
	for _, u := range users {
		if u.UserID == oldUserID {
			continue
		}
		if u.UserID == pr.AuthorID {
			continue
		}
		alreadyAssigned := false
		for _, r := range pr.AssignedReviewers {
			if r == u.UserID {
				alreadyAssigned = true
				break
			}
		}
		if alreadyAssigned {
			continue
		}
		candidates = append(candidates, u)
	}

	if len(candidates) == 0 {
		return nil, "", domain.NewError(domain.ErrorNoCandidate, "no active replacement candidate in team")
	}

	newReviewer := candidates[0].UserID

	for i, r := range pr.AssignedReviewers {
		if r == oldUserID {
			pr.AssignedReviewers[i] = newReviewer
			break
		}
	}

	if err := s.prRepo.Update(ctx, pr); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", domain.NewError(domain.ErrorNotFound, "pull request not found on update: "+prID)
		}
		return nil, "", fmt.Errorf("prRepo.Update: %w", err)
	}

	return pr, newReviewer, nil
}
