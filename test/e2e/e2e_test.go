package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"pr-reviewer-assigment-service/internal/api"
	"pr-reviewer-assigment-service/internal/api/httphandlers"
	"pr-reviewer-assigment-service/internal/application/service"
	"pr-reviewer-assigment-service/internal/infrastructure/postgres"
)

type testDB struct {
	pool      *pgxpool.Pool
	container *tcpg.PostgresContainer
}

func setupTestDB(t *testing.T) *testDB {
	t.Helper()

	ctx := context.Background()

	container, err := tcpg.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		tcpg.WithDatabase("testdb"),
		tcpg.WithUsername("test"),
		tcpg.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	if err := migrateSchema(ctx, pool); err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return &testDB{
		pool:      pool,
		container: container,
	}
}

func teardownTestDB(t *testing.T, db *testDB) {
	t.Helper()
	ctx := context.Background()
	db.pool.Close()
	if err := db.container.Terminate(ctx); err != nil {
		t.Fatalf("failed to terminate container: %v", err)
	}
}

func migrateSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		DROP TABLE IF EXISTS pull_requests;
		DROP TABLE IF EXISTS users;
		DROP TABLE IF EXISTS teams;

		CREATE TABLE teams (
			team_name TEXT PRIMARY KEY
		);

		CREATE TABLE users (
			user_id   TEXT PRIMARY KEY,
			username  TEXT NOT NULL,
			team_name TEXT NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			CONSTRAINT fk_users_team
				FOREIGN KEY (team_name)
				REFERENCES teams(team_name)
				ON UPDATE CASCADE
				ON DELETE RESTRICT
		);

		CREATE TABLE pull_requests (
			pull_request_id     TEXT PRIMARY KEY,
			pull_request_name   TEXT        NOT NULL,
			author_id           TEXT        NOT NULL,
			status              TEXT        NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
			assigned_reviewers  TEXT[]      NOT NULL DEFAULT '{}',
			created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			merged_at           TIMESTAMPTZ NULL,
			CONSTRAINT fk_pull_requests_author
				FOREIGN KEY (author_id)
				REFERENCES users(user_id)
				ON UPDATE CASCADE
				ON DELETE RESTRICT,
			CONSTRAINT assigned_reviewers_max_2
				CHECK (cardinality(assigned_reviewers) <= 2)
		);
	`)
	return err
}

func newTestServer(t *testing.T, db *testDB) *httptest.Server {
	t.Helper()

	userRepo := postgres.NewUserDb(db.pool)
	teamRepo := postgres.NewTeamDb(db.pool)
	prRepo := postgres.NewPullRequestDb(db.pool)

	teamService := service.NewTeamService(userRepo, teamRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo)
	statsService := service.NewStatsService(prRepo)

	teamHandlers := httphandlers.NewTeamHandlers(teamService)
	userHandlers := httphandlers.NewUserHandlers(userService)
	prHandlers := httphandlers.NewPullRequestHandlers(prService)
	statsHandlers := httphandlers.NewStatsHandlers(statsService)

	router := api.NewRouter(
		teamHandlers,
		userHandlers,
		prHandlers,
		statsHandlers,
	)

	return httptest.NewServer(router)
}

type teamAddRequest struct {
	TeamName string `json:"team_name"`
	Members  []struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	} `json:"members"`
}

type teamAddResponse struct {
	Team struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"members"`
	} `json:"team"`
}

type prCreateRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type prCreateResponse struct {
	PR struct {
		PullRequestID     string   `json:"pull_request_id"`
		PullRequestName   string   `json:"pull_request_name"`
		AuthorID          string   `json:"author_id"`
		Status            string   `json:"status"`
		AssignedReviewers []string `json:"assigned_reviewers"`
	} `json:"pr"`
}

type userGetReviewResponse struct {
	UserID       string `json:"user_id"`
	PullRequests []struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
		Status          string `json:"status"`
	} `json:"pull_requests"`
}

func TestE2E_TeamAddAndGet(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ts := newTestServer(t, db)
	defer ts.Close()

	reqBody := teamAddRequest{
		TeamName: "backend",
		Members: []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		}{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	resp, err := http.Post(ts.URL+"/team/add", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("POST /team/add: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var addResp teamAddResponse
	if err := json.NewDecoder(resp.Body).Decode(&addResp); err != nil {
		t.Fatalf("decode addResp: %v", err)
	}

	if addResp.Team.TeamName != "backend" {
		t.Fatalf("unexpected team_name: %s", addResp.Team.TeamName)
	}
	if len(addResp.Team.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(addResp.Team.Members))
	}

	getResp, err := http.Get(ts.URL + "/team/get?team_name=backend")
	if err != nil {
		t.Fatalf("GET /team/get: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", getResp.StatusCode)
	}

	var getBody struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"members"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&getBody); err != nil {
		t.Fatalf("decode getBody: %v", err)
	}

	if getBody.TeamName != "backend" {
		t.Fatalf("unexpected team_name in get: %s", getBody.TeamName)
	}
	if len(getBody.Members) != 2 {
		t.Fatalf("expected 2 members in get, got %d", len(getBody.Members))
	}
}

func TestE2E_FullFlow_CreatePR_And_GetUserReview(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	ts := newTestServer(t, db)
	defer ts.Close()

	teamReq := teamAddRequest{
		TeamName: "backend",
		Members: []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		}{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
			{UserID: "u3", Username: "Charlie", IsActive: true},
		},
	}
	teamBody, _ := json.Marshal(teamReq)

	resp, err := http.Post(ts.URL+"/team/add", "application/json", bytes.NewReader(teamBody))
	if err != nil {
		t.Fatalf("POST /team/add: %v", err)
	}
	resp.Body.Close()

	prReq := prCreateRequest{
		PullRequestID:   "pr-1001",
		PullRequestName: "Add search",
		AuthorID:        "u1",
	}
	prBody, _ := json.Marshal(prReq)

	prResp, err := http.Post(ts.URL+"/pullRequest/create", "application/json", bytes.NewReader(prBody))
	if err != nil {
		t.Fatalf("POST /pullRequest/create: %v", err)
	}
	defer prResp.Body.Close()

	if prResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", prResp.StatusCode)
	}

	var prCreate prCreateResponse
	if err := json.NewDecoder(prResp.Body).Decode(&prCreate); err != nil {
		t.Fatalf("decode prCreate: %v", err)
	}

	if prCreate.PR.PullRequestID != "pr-1001" {
		t.Fatalf("unexpected pr id: %s", prCreate.PR.PullRequestID)
	}
	if prCreate.PR.AuthorID != "u1" {
		t.Fatalf("unexpected author_id: %s", prCreate.PR.AuthorID)
	}
	if len(prCreate.PR.AssignedReviewers) == 0 {
		t.Fatalf("expected at least 1 reviewer assigned")
	}

	reviewerID := prCreate.PR.AssignedReviewers[0]

	getReviewResp, err := http.Get(ts.URL + "/users/getReview?user_id=" + reviewerID)
	if err != nil {
		t.Fatalf("GET /users/getReview: %v", err)
	}
	defer getReviewResp.Body.Close()

	if getReviewResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", getReviewResp.StatusCode)
	}

	var reviewBody userGetReviewResponse
	if err := json.NewDecoder(getReviewResp.Body).Decode(&reviewBody); err != nil {
		t.Fatalf("decode reviewBody: %v", err)
	}

	if reviewBody.UserID != reviewerID {
		t.Fatalf("unexpected user_id in response: %s", reviewBody.UserID)
	}
	if len(reviewBody.PullRequests) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(reviewBody.PullRequests))
	}
	if reviewBody.PullRequests[0].PullRequestID != "pr-1001" {
		t.Fatalf("unexpected PR id in review list: %s", reviewBody.PullRequests[0].PullRequestID)
	}
}
