package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage/postgresql"
)

type Router struct {
	log     *slog.Logger
	storage *postgresql.Storage
}

func notI(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not implemented yet"))
}

func New(log *slog.Logger, storage *postgresql.Storage) http.Handler {
	// Init router
	router := chi.NewRouter()

	// Init middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	//Router
	// Teams
	router.Route("/team", func(team chi.Router) {
		team.Post("/add", notI)
		team.Get("/get", notI)
	})
	// Users
	router.Route("/users", func(users chi.Router) {
		users.Post("/setIsActive", notI)
		users.Get("/getReview", notI)
	})
	// PullRequests
	router.Route("/pullRequest", func(pullRequest chi.Router) {
		pullRequest.Post("/create", notI)
		pullRequest.Post("/merge", notI)
		pullRequest.Post("/reassign", notI)
	})
	return router
}
