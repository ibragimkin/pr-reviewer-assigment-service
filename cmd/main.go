package main

import (
	"context"
	"log"
	"net/http"
	"pr-reviewer-assigment-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"

	"pr-reviewer-assigment-service/internal/api"
	"pr-reviewer-assigment-service/internal/api/httphandlers"
	"pr-reviewer-assigment-service/internal/application/service"
	"pr-reviewer-assigment-service/internal/infrastructure/postgres"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	// repos
	userRepo := postgres.NewUserDb(pool)
	teamRepo := postgres.NewTeamDb(pool)
	prRepo := postgres.NewPullRequestDb(pool)

	// services
	teamService := service.NewTeamService(userRepo, teamRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo)

	// handlers
	teamHandlers := httphandlers.NewTeamHandlers(teamService)
	userHandlers := httphandlers.NewUserHandlers(userService)
	prHandlers := httphandlers.NewPullRequestHandlers(prService)

	// router
	handler := api.NewRouter(teamHandlers, userHandlers, prHandlers)

	log.Println("listening on " + cfg.HttpPort)
	if err := http.ListenAndServe(":"+cfg.HttpPort, handler); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
