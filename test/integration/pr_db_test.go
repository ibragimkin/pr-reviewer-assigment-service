package integration_test

import (
	"context"
	"pr-reviewer-assigment-service/internal/domain"
	pg "pr-reviewer-assigment-service/internal/infrastructure/postgres"
	"testing"
	"time"
)

func TestPullRequestDb_Create_GetByID_And_Stats(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := pg.NewUserDb(db.Pool)
	prRepo := pg.NewPullRequestDb(db.Pool)

	_, err := db.Pool.Exec(ctx, `INSERT INTO teams (team_name) VALUES ('backend')`)
	if err != nil {
		t.Fatalf("insert team: %v", err)
	}

	if err := userRepo.BulkUpsert(ctx, []domain.User{
		{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true},
		{UserID: "u3", Username: "Charlie", TeamName: "backend", IsActive: true},
	}); err != nil {
		t.Fatalf("BulkUpsert users: %v", err)
	}

	now := time.Now().UTC()

	pr1 := &domain.PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            "OPEN",
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         &now,
	}
	pr2 := &domain.PullRequest{
		PullRequestID:     "pr-2",
		PullRequestName:   "Fix bug",
		AuthorID:          "u2",
		Status:            "OPEN",
		AssignedReviewers: []string{"u2"},
		CreatedAt:         &now,
	}

	if err := prRepo.Create(ctx, pr1); err != nil {
		t.Fatalf("Create pr1: %v", err)
	}
	if err := prRepo.Create(ctx, pr2); err != nil {
		t.Fatalf("Create pr2: %v", err)
	}

	got, err := prRepo.GetByID(ctx, "pr-1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.PullRequestName != "Add feature" {
		t.Fatalf("unexpected pr1 name: %s", got.PullRequestName)
	}

	stats, err := prRepo.GetReviewerStats(ctx)
	if err != nil {
		t.Fatalf("GetReviewerStats: %v", err)
	}

	var u2Count, u3Count int
	for _, s := range stats {
		if s.UserID == "u2" {
			u2Count = s.ReviewCount
		}
		if s.UserID == "u3" {
			u3Count = s.ReviewCount
		}
	}

	if u2Count != 2 {
		t.Fatalf("expected u2 review_count=2, got %d", u2Count)
	}
	if u3Count != 1 {
		t.Fatalf("expected u3 review_count=1, got %d", u3Count)
	}
}
