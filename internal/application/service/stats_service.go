package service

import (
	"context"
	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/domain"
)

type StatsService struct {
	prRepo repository.PullRequestRepository
}

func NewStatsService(prRepo repository.PullRequestRepository) *StatsService {
	return &StatsService{prRepo: prRepo}
}

func (s *StatsService) GetReviewerStats(ctx context.Context) ([]domain.ReviewerStat, error) {
	return s.prRepo.GetReviewerStats(ctx)
}
