package router

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/domain"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/transport"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger/sl"
)

func (router *Router) TPOSTAdd(w http.ResponseWriter, r *http.Request) {
	type request struct {
		TeamName string `json:"team_name" validate:"required"`
		Members  []struct {
			UserID   string `json:"user_id" validate:"required"`
			Username string `json:"username" validate:"required"`
			IsActive bool   `json:"is_active" validate:"required"`
		}
	}
	type response struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		}
	}

	// Декодирование и валидация request
	var req request
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		router.log.Error("failed to decode request", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.BAD_REQUEST,
			Message: "failed to decode request",
		})
		return
	}
	if err := validator.New().Struct(req); err != nil {
		router.log.Error("failed to validate request", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.BAD_REQUEST,
			Message: "failed to validate request",
		})
		return
	}

	// Создаем юзеров в объект меж сервисами
	users := make([]domain.User, 0)
	for _, user := range req.Members {
		users = append(users, domain.User{ID: user.UserID, Name: user.Username, IsActive: user.IsActive})
	}

	err := router.storage.CreateTeamWithUser(req.TeamName, users)
	if err != nil {
		if errors.Is(err, storage.ErrTeamAlreadyExists) {
			router.log.Error("failed to create team", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.TEAM_EXISTS,
				Message: "team_name already exists",
			})
			return
		}
		router.log.Error("failed to create team", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, response{
		TeamName: req.TeamName,
		Members: []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		}(req.Members),
	})
}

func (router *Router) TGET(w http.ResponseWriter, r *http.Request) {
	type respMembers struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	}
	type response struct {
		TeamName string `json:"team_name"`
		Members  []respMembers
	}

	teamName := r.URL.Query().Get("team_name")

	infoTeam, err := router.storage.GetTeam(teamName)
	if err != nil {
		if errors.Is(err, storage.ErrTeamNotFound) {
			router.log.Error("failed to find team", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		router.log.Error("failed to get team", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: err.Error(),
		})
		return
	}

	members := make([]respMembers, 0, len(infoTeam.Users))
	for _, user := range infoTeam.Users {
		members = append(members, respMembers{
			UserID:   user.ID,
			Username: user.Name,
			IsActive: user.IsActive,
		})
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response{
		TeamName: teamName,
		Members:  members,
	})
}

func (router *Router) DeactivateTeamUsers(w http.ResponseWriter, r *http.Request) {
	type request struct {
		TeamName string `json:"team_name" validate:"required"`
	}
	type response struct {
		TeamName        string `json:"team_name"`
		DeactivateCount int    `json:"deactivate_count"`
	}

	// Декодирование и валидация request
	var req request
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		router.log.Error("failed to decode request", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.BAD_REQUEST,
			Message: "failed to decode request",
		})
		return
	}
	if err := validator.New().Struct(req); err != nil {
		router.log.Error("failed to validate request", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.BAD_REQUEST,
			Message: "failed to validate request",
		})
		return
	}

	count, err := router.storage.DeactivateTeamUsers(req.TeamName)
	if err != nil {
		if errors.Is(err, storage.ErrTeamNotFound) {
			router.log.Error("failed to deactivate team", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		router.log.Error("failed to deactivate team users", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed to deactivate team users",
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response{
		TeamName:        req.TeamName,
		DeactivateCount: count,
	})
}
