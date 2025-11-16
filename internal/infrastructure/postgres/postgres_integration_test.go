package postgres_test

import (
	"context"
	_ "fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tc "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testDB struct {
	Pool      *pgxpool.Pool
	Container *tc.PostgresContainer
}

func setupTestDB(t *testing.T) *testDB {
	t.Helper()

	ctx := context.Background()

	container, err := tc.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		tc.WithDatabase("testdb"),
		tc.WithUsername("test"),
		tc.WithPassword("test"),
		tc.WithInitScripts(),
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

	if err := migrateTestSchema(ctx, pool); err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return &testDB{
		Pool:      pool,
		Container: container,
	}
}

func teardownTestDB(t *testing.T, db *testDB) {
	t.Helper()
	ctx := context.Background()
	db.Pool.Close()
	if err := db.Container.Terminate(ctx); err != nil {
		t.Fatalf("failed to terminate container: %v", err)
	}
}

func migrateTestSchema(ctx context.Context, db *pgxpool.Pool) error {
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
