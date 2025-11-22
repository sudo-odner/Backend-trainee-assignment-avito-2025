package router

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
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

	// TODO: Create/update user
	// TODO: Создание и добавление пользователей

	// TODO: Ответ
}

func (router *Router) TGET(w http.ResponseWriter, r *http.Request) {
	type response struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		}
	}

	teamName := r.URL.Query().Get("team_name")

	// TODO: Get team data

	// TODOЖ Ответ
}
