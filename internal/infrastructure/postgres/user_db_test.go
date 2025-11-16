package postgres_test

import (
	"context"
	"pr-reviewer-assigment-service/internal/domain"
	pg "pr-reviewer-assigment-service/internal/infrastructure/postgres"
	"testing"
)

func TestUserDb_BulkUpsert_And_GetByID(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := pg.NewUserDb(db.Pool)

	_, err := db.Pool.Exec(ctx, `INSERT INTO teams (team_name) VALUES ('backend')`)
	if err != nil {
		t.Fatalf("insert team: %v", err)
	}

	users := []domain.User{
		{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true},
		{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: false},
	}

	if err := userRepo.BulkUpsert(ctx, users); err != nil {
		t.Fatalf("BulkUpsert: %v", err)
	}

	u1, err := userRepo.GetByID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if u1.Username != "Alice" || u1.TeamName != "backend" || !u1.IsActive {
		t.Fatalf("unexpected u1: %+v", u1)
	}

	updated, err := userRepo.SetActive(ctx, "u2", true)
	if err != nil {
		t.Fatalf("SetActive: %v", err)
	}
	if !updated.IsActive {
		t.Fatalf("expected u2 IsActive=true, got %v", updated.IsActive)
	}
}
