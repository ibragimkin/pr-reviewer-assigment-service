package api

import (
	_ "embed"
	"github.com/go-chi/chi/v5"
	"github.com/go-openapi/runtime/middleware"
	"net/http"
	"pr-reviewer-assigment-service/internal/api/httphandlers"
)

//go:embed docs/openapi.yml
var OpenAPISpec []byte

// NewRouter принимает все группы хендлеров и возвращает готовый http.Handler.
func NewRouter(
	teamHandlers *httphandlers.TeamHandlers,
	userHandlers *httphandlers.UserHandlers,
	prHandlers *httphandlers.PullRequestHandlers,
) http.Handler {
	r := chi.NewRouter()
	r.Get("/swagger/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(OpenAPISpec)
	})

	swaggerOpts := middleware.SwaggerUIOpts{
		SpecURL: "/swagger/openapi.yml",
		Path:    "/swagger",
	}

	swaggerHandler := middleware.SwaggerUI(swaggerOpts, nil)

	r.Handle("/swagger", swaggerHandler)
	r.Handle("/swagger/*", swaggerHandler)

	r.Post("/team/add", teamHandlers.Add)
	r.Get("/team/get", teamHandlers.Get)

	r.Post("/users/setIsActive", userHandlers.SetIsActive)
	r.Get("/users/getReview", userHandlers.GetReview)

	r.Post("/pullRequest/create", prHandlers.Create)
	r.Post("/pullRequest/merge", prHandlers.Merge)
	r.Post("/pullRequest/reassign", prHandlers.Reassign)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return r
}
