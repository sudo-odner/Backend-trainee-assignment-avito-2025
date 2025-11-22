package router

import (
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/transport"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger/sl"
)

func (router *Router) UserPOSTSetIsActivate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	type response struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
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

	// TODO: Update User in DB
	// TODO: Update Get TeamName for User

	// TODO: Ответ
}

func (router *Router) UserGETGetReview(w http.ResponseWriter, r *http.Request) {
	type response struct {
		UserID       string `json:"user_id"`
		PullRequests []struct {
			PullRequestID   string `json:"pull_request_id"`
			PullRequestName string `json:"pull_request_name"`
			AuthorID        string `json:"author_id"`
			Status          string `json:"status"`
		}
	}

	userID := r.URL.Query().Get("user_id")

	// TODO: Получить все ПР для данного юзера

	// TODO: Ответ сервера
}
