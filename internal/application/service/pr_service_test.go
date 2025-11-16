package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/application/service"
	"pr-reviewer-assigment-service/internal/domain"
)

type mockPRRepo struct {
	data map[string]domain.PullRequest
}

func newMockPRRepo() *mockPRRepo {
	return &mockPRRepo{
		data: make(map[string]domain.PullRequest),
	}
}

func (m *mockPRRepo) Create(ctx context.Context, pr *domain.PullRequest) error {
	if _, ok := m.data[pr.PullRequestID]; ok {
		return repository.ErrAlreadyExists
	}
	cp := *pr
	m.data[pr.PullRequestID] = cp
	return nil
}

func (m *mockPRRepo) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	pr, ok := m.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cp := pr
	return &cp, nil
}

func (m *mockPRRepo) Update(ctx context.Context, pr *domain.PullRequest) error {
	if _, ok := m.data[pr.PullRequestID]; !ok {
		return repository.ErrNotFound
	}
	cp := *pr
	m.data[pr.PullRequestID] = cp

	return nil
}

func (m *mockPRRepo) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	result := []domain.PullRequestShort{}

	for _, pr := range m.data {
		for _, r := range pr.AssignedReviewers {
			if r == reviewerID {
				result = append(result, domain.PullRequestShort{
					PullRequestID:   pr.PullRequestID,
					PullRequestName: pr.PullRequestName,
					AuthorID:        pr.AuthorID,
					Status:          pr.Status,
				})
				break
			}
		}
	}

	return result, nil
}

func TestPullRequestService_Create_SuccessTwoReviewers(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	userRepo.data["u1"] = domain.User{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true}
	userRepo.data["u2"] = domain.User{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true}
	userRepo.data["u3"] = domain.User{UserID: "u3", Username: "Charlie", TeamName: "backend", IsActive: true}
	teamRepo.data["backend"] = domain.Team{TeamName: "backend"}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	pr, err := svc.Create(ctx, "pr-1", "Add feature", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pr.PullRequestID != "pr-1" {
		t.Fatalf("unexpected pr id: %s", pr.PullRequestID)
	}
	if pr.Status != "OPEN" {
		t.Fatalf("unexpected status: %s", pr.Status)
	}
	if len(pr.AssignedReviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(pr.AssignedReviewers))
	}
	for _, r := range pr.AssignedReviewers {
		if r == "u1" {
			t.Fatalf("author was assigned as reviewer")
		}
	}
	if pr.CreatedAt == nil {
		t.Fatalf("expected CreatedAt to be set")
	}
}

func TestPullRequestService_Create_SuccessOneReviewer(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	userRepo.data["u1"] = domain.User{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true}
	userRepo.data["u2"] = domain.User{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true}
	teamRepo.data["backend"] = domain.Team{TeamName: "backend"}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	pr, err := svc.Create(ctx, "pr-2", "Fix bug", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pr.AssignedReviewers) != 1 {
		t.Fatalf("expected 1 reviewer, got %d", len(pr.AssignedReviewers))
	}
	if pr.AssignedReviewers[0] != "u2" {
		t.Fatalf("unexpected reviewer: %s", pr.AssignedReviewers[0])
	}
}

func TestPullRequestService_Create_NoReviewers(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	userRepo.data["u1"] = domain.User{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true}
	teamRepo.data["backend"] = domain.Team{TeamName: "backend"}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	pr, err := svc.Create(ctx, "pr-3", "Doc change", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pr.AssignedReviewers) != 0 {
		t.Fatalf("expected 0 reviewers, got %d", len(pr.AssignedReviewers))
	}
}

func TestPullRequestService_Create_AuthorNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, err := svc.Create(ctx, "pr-4", "Add feature", "u-missing")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", dErr.Code)
	}
}

func TestPullRequestService_Create_TeamNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	userRepo.data["u1"] = domain.User{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, err := svc.Create(ctx, "pr-5", "Add feature", "u1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", dErr.Code)
	}
}

func TestPullRequestService_Create_PRAlreadyExists(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	userRepo.data["u1"] = domain.User{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true}
	userRepo.data["u2"] = domain.User{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true}
	teamRepo.data["backend"] = domain.Team{TeamName: "backend"}

	prRepo.data["pr-6"] = domain.PullRequest{
		PullRequestID:   "pr-6",
		PullRequestName: "Existing",
		AuthorID:        "u1",
		Status:          "OPEN",
	}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, err := svc.Create(ctx, "pr-6", "Duplicate", "u1")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorPRExists {
		t.Fatalf("expected PR_EXISTS, got %s", dErr.Code)
	}
}

func TestPullRequestService_Merge_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	now := time.Now().UTC()
	prRepo.data["pr-7"] = domain.PullRequest{
		PullRequestID:   "pr-7",
		PullRequestName: "To merge",
		AuthorID:        "u1",
		Status:          "OPEN",
		CreatedAt:       &now,
	}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	pr, err := svc.Merge(ctx, "pr-7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pr.Status != "MERGED" {
		t.Fatalf("expected MERGED, got %s", pr.Status)
	}
	if pr.MergedAt == nil {
		t.Fatalf("expected MergedAt to be set")
	}
	stored, _ := prRepo.GetByID(ctx, "pr-7")
	if stored.Status != pr.Status {
		t.Fatalf("repo not updated")
	}
}

func TestPullRequestService_Merge_NotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, err := svc.Merge(ctx, "no-pr")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", dErr.Code)
	}
}

func TestPullRequestService_Reassign_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	userRepo.data["u2"] = domain.User{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true}
	userRepo.data["u3"] = domain.User{UserID: "u3", Username: "Charlie", TeamName: "backend", IsActive: true}
	userRepo.data["u4"] = domain.User{UserID: "u4", Username: "Dave", TeamName: "backend", IsActive: true}
	teamRepo.data["backend"] = domain.Team{TeamName: "backend"}

	prRepo.data["pr-8"] = domain.PullRequest{
		PullRequestID:     "pr-8",
		PullRequestName:   "Reassign test",
		AuthorID:          "u1",
		Status:            "OPEN",
		AssignedReviewers: []string{"u2", "u3"},
	}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	pr, replacedBy, err := svc.Reassign(ctx, "pr-8", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if replacedBy == "" {
		t.Fatalf("expected non-empty replacedBy")
	}
	if replacedBy == "u2" {
		t.Fatalf("expected new reviewer different from old")
	}
	foundOld := false
	foundNew := false
	for _, r := range pr.AssignedReviewers {
		if r == "u2" {
			foundOld = true
		}
		if r == replacedBy {
			foundNew = true
		}
	}
	if foundOld {
		t.Fatalf("old reviewer still present")
	}
	if !foundNew {
		t.Fatalf("new reviewer not present")
	}
}

func TestPullRequestService_Reassign_PRNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, _, err := svc.Reassign(ctx, "no-pr", "u2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", dErr.Code)
	}
}

func TestPullRequestService_Reassign_PRMerged(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	prRepo.data["pr-9"] = domain.PullRequest{
		PullRequestID:     "pr-9",
		PullRequestName:   "Merged",
		AuthorID:          "u1",
		Status:            "MERGED",
		AssignedReviewers: []string{"u2"},
	}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, _, err := svc.Reassign(ctx, "pr-9", "u2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorPRMerged {
		t.Fatalf("expected PR_MERGED, got %s", dErr.Code)
	}
}

func TestPullRequestService_Reassign_OldUserNotAssigned(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	prRepo.data["pr-10"] = domain.PullRequest{
		PullRequestID:     "pr-10",
		PullRequestName:   "Not assigned",
		AuthorID:          "u1",
		Status:            "OPEN",
		AssignedReviewers: []string{"u3"},
	}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, _, err := svc.Reassign(ctx, "pr-10", "u2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorNotAssigned {
		t.Fatalf("expected NOT_ASSIGNED, got %s", dErr.Code)
	}
}

func TestPullRequestService_Reassign_ReviewerNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	prRepo.data["pr-11"] = domain.PullRequest{
		PullRequestID:     "pr-11",
		PullRequestName:   "Reviewer missing",
		AuthorID:          "u1",
		Status:            "OPEN",
		AssignedReviewers: []string{"u2"},
	}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, _, err := svc.Reassign(ctx, "pr-11", "u2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorNotFound {
		t.Fatalf("expected NOT_FOUND, got %s", dErr.Code)
	}
}

func TestPullRequestService_Reassign_NoCandidate(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()
	prRepo := newMockPRRepo()

	userRepo.data["u2"] = domain.User{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true}
	teamRepo.data["backend"] = domain.Team{TeamName: "backend"}

	prRepo.data["pr-12"] = domain.PullRequest{
		PullRequestID:     "pr-12",
		PullRequestName:   "No candidate",
		AuthorID:          "u1",
		Status:            "OPEN",
		AssignedReviewers: []string{"u2"},
	}

	svc := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	_, _, err := svc.Reassign(ctx, "pr-12", "u2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}
	if dErr.Code != domain.ErrorNoCandidate {
		t.Fatalf("expected NO_CANDIDATE, got %s", dErr.Code)
	}
}
