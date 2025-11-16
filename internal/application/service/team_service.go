package service

import (
	"context"
	"errors"
	"fmt"
	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/domain"
)

type TeamService struct {
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
}

func NewTeamService(userRepository repository.UserRepository, teamRepository repository.TeamRepository) *TeamService {
	return &TeamService{userRepo: userRepository, teamRepo: teamRepository}
}

func (t *TeamService) Add(ctx context.Context, teamName string, members []domain.TeamMember) (*domain.Team, error) {
	existing, err := t.teamRepo.GetByName(ctx, teamName)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("teamRepo.GetByName: %w", err)
	}

	if existing != nil {
		return nil, domain.NewError(domain.ErrorTeamExists,
			fmt.Sprintf("team %s already exists", teamName))
	}

	team := &domain.Team{
		TeamName: teamName,
		Members:  members,
	}

	if err := t.teamRepo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("teamRepo.Create: %w", err)
	}

	users := make([]domain.User, 0, len(members))
	for _, m := range members {
		users = append(users, domain.User{
			UserID:   m.UserID,
			Username: m.Username,
			TeamName: teamName,
			IsActive: m.IsActive,
		})
	}

	if err := t.userRepo.BulkUpsert(ctx, users); err != nil {
		return nil, fmt.Errorf("userRepo.BulkUpsert: %w", err)
	}
	return team, nil
}

func (t *TeamService) Get(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := t.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.NewError(domain.ErrorNotFound, "team not found: "+teamName)
		}
		return nil, fmt.Errorf("teamRepo.GetByName: %w", err)
	}

	return team, nil
}

func (t *TeamService) DeactivateMembers(ctx context.Context, teamName string) error {
	team, err := t.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.NewError(domain.ErrorNotFound, "team not found: "+teamName)
		}
	}
	for _, member := range team.Members {
		_, err := t.userRepo.SetActive(ctx, member.UserID, false)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				continue
			}
			return fmt.Errorf("userRepo.SetActive: %w", err)
		}
	}
	return nil
}
