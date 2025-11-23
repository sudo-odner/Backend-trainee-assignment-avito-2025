package router

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/storage"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/transport"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger/sl"
)

func (router *Router) UserPOSTSetIsActivate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	type responseUser struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
	}

	type response struct {
		User responseUser `json:"user"`
	}

	// Валидация и декодирование запроса
	var req request
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		router.log.Error("failed decode request", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.BAD_REQUEST,
			Message: "failed decode request",
		})
		return
	}

	if err := validator.New().Struct(req); err != nil {
		router.log.Error("failed validate request", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.BAD_REQUEST,
			Message: "failed validate request",
		})
		return
	}

	// Получение информации о пользоватлеле
	user, err := router.storage.GetUserByID(req.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			router.log.Error("user not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		router.log.Error("failed get user by id", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed get user by id",
		})
		return
	}

	// Установть флаг активности
	if err := router.storage.SetUserIsActive(req.UserID, req.IsActive); err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			router.log.Error("user not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		router.log.Error("failed set user is_active", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed set user is_active",
		})
		return
	}
	// Получить команду пользователя
	teamName, err := router.storage.GetUserTeamByID(user.ID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			router.log.Error("user not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		router.log.Error("failed get user by id", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed get user by id",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response{
		User: responseUser{
			UserID:   user.ID,
			Username: user.Name,
			TeamName: teamName,
			IsActive: req.IsActive,
		},
	})
	return
}

func (router *Router) UserGETGetReview(w http.ResponseWriter, r *http.Request) {
	type responsePR struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
		Status          string `json:"status"`
	}
	type response struct {
		UserID       string       `json:"user_id"`
		PullRequests []responsePR `json:"pull_requests"`
	}

	userID := r.URL.Query().Get("user_id")

	prs, err := router.storage.GetUserPRsByID(userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			router.log.Error("user not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		router.log.Error("failed get user by id", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed get user by id",
		})
		return
	}
	responsePRs := make([]responsePR, 0, len(prs))
	for _, pr := range prs {
		responsePRs = append(responsePRs, responsePR{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.Author.ID,
			Status:          pr.Status,
		})
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response{
		UserID:       userID,
		PullRequests: responsePRs,
	})
	return
}
