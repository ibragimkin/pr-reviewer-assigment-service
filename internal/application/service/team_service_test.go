package service_test

import (
	"context"
	"errors"
	"testing"

	"pr-reviewer-assigment-service/internal/application/repository"
	"pr-reviewer-assigment-service/internal/application/service"
	"pr-reviewer-assigment-service/internal/domain"
)

type mockTeamRepo struct {
	data map[string]domain.Team
}

func newMockTeamRepo() *mockTeamRepo {
	return &mockTeamRepo{
		data: make(map[string]domain.Team),
	}
}

func (m *mockTeamRepo) Create(ctx context.Context, team *domain.Team) error {
	if _, ok := m.data[team.TeamName]; ok {
		return repository.ErrAlreadyExists
	}

	cp := *team
	if team.Members != nil {
		cp.Members = append([]domain.TeamMember(nil), team.Members...)
	}

	m.data[team.TeamName] = cp
	return nil
}

func (m *mockTeamRepo) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	t, ok := m.data[teamName]
	if !ok {
		return nil, repository.ErrNotFound
	}

	cp := t
	if t.Members != nil {
		cp.Members = append([]domain.TeamMember(nil), t.Members...)
	}

	return &cp, nil
}

func TestTeamService_Add_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()

	svc := service.NewTeamService(userRepo, teamRepo)

	members := []domain.TeamMember{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: false},
	}

	team, err := svc.Add(ctx, "backend", members)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team == nil {
		t.Fatalf("expected team, got nil")
	}

	if team.TeamName != "backend" {
		t.Fatalf("expected team name backend, got %s", team.TeamName)
	}

	if len(team.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(team.Members))
	}

	storedTeam, err := teamRepo.GetByName(ctx, "backend")
	if err != nil {
		t.Fatalf("team not stored in repo: %v", err)
	}
	if len(storedTeam.Members) != 2 {
		t.Fatalf("expected 2 members in repo, got %d", len(storedTeam.Members))
	}

	u1, ok := userRepo.data["u1"]
	if !ok {
		t.Fatalf("user u1 not upserted")
	}
	if u1.TeamName != "backend" || u1.Username != "Alice" || u1.IsActive != true {
		t.Fatalf("user u1 has wrong data: %+v", u1)
	}

	u2, ok := userRepo.data["u2"]
	if !ok {
		t.Fatalf("user u2 not upserted")
	}
	if u2.TeamName != "backend" || u2.Username != "Bob" || u2.IsActive != false {
		t.Fatalf("user u2 has wrong data: %+v", u2)
	}
}

func TestTeamService_Add_TeamAlreadyExists(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()

	teamRepo.data["backend"] = domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	svc := service.NewTeamService(userRepo, teamRepo)

	members := []domain.TeamMember{
		{UserID: "u2", Username: "Bob", IsActive: true},
	}

	_, err := svc.Add(ctx, "backend", members)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}

	if dErr.Code != domain.ErrorTeamExists {
		t.Fatalf("expected error code TEAM_EXISTS, got %s", dErr.Code)
	}

	if len(userRepo.data) != 0 {
		t.Fatalf("expected no users upserted, got %d", len(userRepo.data))
	}
}

func TestTeamService_Get_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()

	teamRepo.data["backend"] = domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	}

	svc := service.NewTeamService(userRepo, teamRepo)

	team, err := svc.Get(ctx, "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if team == nil {
		t.Fatalf("expected team, got nil")
	}

	if team.TeamName != "backend" {
		t.Fatalf("expected team_name backend, got %s", team.TeamName)
	}

	if len(team.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(team.Members))
	}
}

func TestTeamService_Get_NotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := newMockUserRepo()
	teamRepo := newMockTeamRepo()

	svc := service.NewTeamService(userRepo, teamRepo)

	_, err := svc.Get(ctx, "unknown")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var dErr *domain.Error
	if !errors.As(err, &dErr) {
		t.Fatalf("expected domain.Error, got %T: %v", err, err)
	}

	if dErr.Code != domain.ErrorNotFound {
		t.Fatalf("expected error code NOT_FOUND, got %s", dErr.Code)
	}
}
