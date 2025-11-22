package router

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/transport"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger/sl"
)

func (router *Router) PRPOSTCreate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		PullRequestID   string `json:"pull_Request_id" validate:"required"`
		PullRequestName int    `json:"pull_Request_number" validate:"required"`
		AuthorName      string `json:"author_name" validate:"required"`
	}
	type response struct {
		PullRequestId     string   `json:"pull_request_id"`
		PullRequestName   string   `json:"pull_request_name"`
		AuthorName        string   `json:"author_name"`
		Status            string   `json:"status"`
		AssignedReviewers []string `json:"assigned_reviewers"`
	}

	// Декодирование и валидация запроса
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

	// TODO: Создание PR (Прикрепление до 2 ревюеров)
	// TODO: Генерация ответа
}

func (router *Router) PRPOSTMerge(w http.ResponseWriter, r *http.Request) {
	type request struct {
		PullRequestID string `json:"pull_request_id" validate:"required"`
	}
	type response struct {
		Pr struct {
			PullRequestId   string   `json:"pull_request_id"`
			PullRequestName string   `json:"pull_request_name"`
			AuthorID        string   `json:"author_id"`
			Status          string   `json:"status"`
			AssignedReviews []string `json:"assigned_reviewers"`
			MergedAt        string   `json:"merged_at"`
		} `json:"pr"`
	}

	// decode and validation
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

	// TODO: Отметить PR как MERGED (если до этого уже MERGED, время тоже самое(идемпотентная операция)
	// TODO: Ответ
}

func (router *Router) PRPOSTReassign(w http.ResponseWriter, r *http.Request) {
	type request struct {
		PullRequestID string `json:"pull_request_id" validate:"required"`
		OldReviewerID string `json:"old_reviewer_id" validate:"required"`
	}
	type response struct {
		Pr struct {
			PullResuestId   string   `json:"pull_request_id"`
			PullRequestName string   `json:"pull_request_name"`
			AuthorName      string   `json:"author_name"`
			Status          string   `json:"status"`
			AssignedReviews []string `json:"assigned_reviewers"`
		} `json:"pr"`
		ReplacedBy string `json:"replaced_by"`
	}
	// Validation and decode
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
	// TODO: Переназначить ревюера на другого из команды (если это возможно)

}
