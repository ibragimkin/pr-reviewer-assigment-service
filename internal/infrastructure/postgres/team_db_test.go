package postgres_test

import (
	"context"
	"pr-reviewer-assigment-service/internal/domain"
	pg "pr-reviewer-assigment-service/internal/infrastructure/postgres"
	"testing"
)

func TestTeamDb_Create_And_GetByName(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	teamRepo := pg.NewTeamDb(db.Pool)

	team := &domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	}

	if err := teamRepo.Create(ctx, team); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES
			('u1', 'Alice', 'backend', true),
			('u2', 'Bob', 'backend', false)
	`)
	if err != nil {
		t.Fatalf("insert users: %v", err)
	}

	got, err := teamRepo.GetByName(ctx, "backend")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}

	if got.TeamName != "backend" {
		t.Fatalf("unexpected team_name: %s", got.TeamName)
	}
	if len(got.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(got.Members))
	}
}
