package service_test

import (
	"context"
	"errors"
	"testing"

	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/application/service"
	"pr-reviewer-assigment-service/internal/domain"
)

type mockUserRepo struct {
	data map[string]domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		data: make(map[string]domain.User),
	}
}

func (m *mockUserRepo) BulkUpsert(ctx context.Context, users []domain.User) error {
	for _, user := range users {
		m.data[user.UserID] = user
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if user, ok := m.data[userID]; ok {
		return &user, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockUserRepo) SetActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	user, ok := m.data[userID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	user.IsActive = active
	m.data[userID] = user
	return &user, nil
}

func (m *mockUserRepo) ListByTeam(ctx context.Context, teamName string, onlyActive bool) ([]domain.User, error) {
	users := make([]domain.User, 0)
	for _, user := range m.data {
		if user.TeamName == teamName {
			users = append(users, user)
		}
	}
	return users, nil
}

func TestUserService_SetIsActive_UserNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	prRepo := newMockPRRepo()

	svc := service.NewUserService(userRepo, prRepo)

	_, err := svc.SetIsActive(ctx, "nope", false)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var derr *domain.Error
	if !errors.As(err, &derr) {
		t.Fatalf("expected domain.Error, got %v", err)
	}

	if derr.Code != domain.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", derr.Code)
	}
}

func TestUserService_SetIsActive_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	prRepo := newMockPRRepo()
	svc := service.NewUserService(userRepo, prRepo)

	userRepo.data["u1"] = domain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	user, err := svc.SetIsActive(ctx, "u1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.IsActive != false {
		t.Fatalf("expected isActive=false, got %v", user.IsActive)
	}

	if userRepo.data["u1"].IsActive != false {
		t.Fatalf("repo not updated")
	}
}

func TestUserService_GetReview_UserNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	prRepo := newMockPRRepo()

	svc := service.NewUserService(userRepo, prRepo)

	_, _, err := svc.GetReview(ctx, "ghost")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var derr *domain.Error
	if !errors.As(err, &derr) {
		t.Fatalf("expected domain.Error, got %v", err)
	}

	if derr.Code != domain.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", derr.Code)
	}
}

func TestUserService_GetReview_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	prRepo := newMockPRRepo()

	svc := service.NewUserService(userRepo, prRepo)

	userRepo.data["u1"] = domain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	prRepo.data["pr1"] = domain.PullRequest{
		PullRequestID:     "pr1",
		PullRequestName:   "Fix bug",
		AuthorID:          "u2",
		Status:            "OPEN",
		AssignedReviewers: []string{"u1"},
	}

	prRepo.data["pr2"] = domain.PullRequest{
		PullRequestID:     "pr2",
		PullRequestName:   "Add login",
		AuthorID:          "u3",
		Status:            "OPEN",
		AssignedReviewers: []string{"u1", "u4"},
	}

	userID, prs, err := svc.GetReview(ctx, "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if userID != "u1" {
		t.Fatalf("expected userID=u1, got %s", userID)
	}

	if len(prs) != 2 {
		t.Fatalf("expected 2 PRs, got %d", len(prs))
	}
}
