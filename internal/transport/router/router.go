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

func New(log *slog.Logger, storage *postgresql.Storage) http.Handler {
	r := Router{
		log:     log,
		storage: storage,
	}
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
		team.Post("/add", r.TPOSTAdd)
		team.Get("/get", r.TGET)
	})
	// Users
	router.Route("/users", func(users chi.Router) {
		users.Post("/setIsActive", r.UserPOSTSetIsActivate)
		users.Get("/getReview", r.UserGETGetReview)
	})
	// PullRequests
	router.Route("/pullRequest", func(pullRequest chi.Router) {
		pullRequest.Post("/create", r.PRPOSTCreate)
		pullRequest.Post("/merge", r.PRPOSTMerge)
		pullRequest.Post("/reassign", r.PRPOSTReassign)
	})
	return router
}
