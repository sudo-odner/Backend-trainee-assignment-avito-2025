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

func (router *Router) PRPOSTCreate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		PullRequestID   string `json:"pull_request_id" validate:"required"`
		PullRequestName string `json:"pull_request_name" validate:"required"`
		AuthorID        string `json:"author_id" validate:"required"`
	}
	type response struct {
		PullRequestID     string   `json:"pull_request_id"`
		PullRequestName   string   `json:"pull_request_name"`
		AuthorID          string   `json:"author_name"`
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

	pr, err := router.storage.CreatePRWithReviewers(req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) || errors.Is(err, storage.ErrTeamNotFound) {
			router.log.Error("Not found user or team", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.BAD_REQUEST,
				Message: "resource not found",
			})
			return
		}
		if errors.Is(err, storage.ErrPRAlreadyExists) {
			router.log.Error("PR already exists", sl.Err(err))
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.PR_EXISTS,
				Message: "PR id already exists",
			})
			return
		}
		router.log.Error("failed to create PR", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed to create PR",
		})
		return
	}
	assignedReviewers := make([]string, 0, len(pr.Reviewers))
	for _, reviewer := range pr.Reviewers {
		assignedReviewers = append(assignedReviewers, reviewer.ID)
	}
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, response{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.Author.ID,
		Status:            pr.Status,
		AssignedReviewers: assignedReviewers,
	})
}

func (router *Router) PRPOSTMerge(w http.ResponseWriter, r *http.Request) {
	type request struct {
		PullRequestID string `json:"pull_request_id" validate:"required"`
	}
	type responsePR struct {
		PullRequestID   string   `json:"pull_request_id"`
		PullRequestName string   `json:"pull_request_name"`
		AuthorID        string   `json:"author_id"`
		Status          string   `json:"status"`
		AssignedReviews []string `json:"assigned_reviewers"`
		MergedAt        string   `json:"merged_at"`
	}
	type response struct {
		PR responsePR `json:"pr"`
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

	// Отметить PR как MERGED (если до этого уже MERGED, время тоже самое(идемпотентная операция)
	err := router.storage.MergePR(req.PullRequestID)
	if err != nil && !errors.Is(err, storage.ErrPRAlreadyMerged) {
		if errors.Is(err, storage.ErrPRNotFound) {
			router.log.Error("PR not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		router.log.Error("failed to soft-merge PR", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed to soft-merge PR",
		})
		return
	}
	pr, err := router.storage.GetPRByID(req.PullRequestID)
	if err != nil {
		if errors.Is(err, storage.ErrPRNotFound) {
			router.log.Error("PR not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		router.log.Error("failed to get PR", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed to get PR",
		})
		return
	}
	reviewers := make([]string, 0, len(pr.Reviewers))
	for _, reviewer := range pr.Reviewers {
		reviewers = append(reviewers, reviewer.ID)
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response{
		PR: responsePR{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.Author.ID,
			Status:          pr.Status,
			AssignedReviews: reviewers,
			MergedAt:        pr.MergedAt.String(),
		},
	})
}

func (router *Router) PRPOSTReassign(w http.ResponseWriter, r *http.Request) {
	type request struct {
		PullRequestID string `json:"pull_request_id" validate:"required"`
		OldReviewerID string `json:"old_reviewer_id" validate:"required"`
	}
	type responsePR struct {
		PullResuestID   string   `json:"pull_request_id"`
		PullRequestName string   `json:"pull_request_name"`
		AuthorID        string   `json:"author_id"`
		Status          string   `json:"status"`
		AssignedReviews []string `json:"assigned_reviewers"`
	}
	type response struct {
		PR         responsePR `json:"pr"`
		ReplacedBy string     `json:"replaced_by"`
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
	pr, newReviewer, err := router.storage.ReassignReviewer(req.PullRequestID, req.OldReviewerID)
	if err != nil {
		if errors.Is(err, storage.ErrPRNotFound) || errors.Is(err, storage.ErrUserNotFound) {
			router.log.Error("PR or user not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_FOUND,
				Message: "resource not found",
			})
			return
		}
		if errors.Is(err, storage.ErrPRAlreadyMerged) {
			router.log.Error("PR already merged", sl.Err(err))
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.PR_MERGED,
				Message: "cannot reassign on merged PR",
			})
			return
		}
		if errors.Is(err, storage.ErrNoCandidate) {
			router.log.Error("PR no candidate", sl.Err(err))
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NO_CANDIDATE,
				Message: "resource not found",
			})
			return
		}
		if errors.Is(err, storage.ErrReviewerNotAssigned) {
			router.log.Error("PR reviewer not assigned", sl.Err(err))
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, transport.ErrResponse{
				Code:    transport.NOT_ASSIGNED,
				Message: "reviewer is not assigned to this PR",
			})
			return
		}
		router.log.Error("failed to reassign PR", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed to reassign PR",
		})
		return
	}
	reviewers := make([]string, 0, len(pr.Reviewers))
	for _, reviewer := range pr.Reviewers {
		reviewers = append(reviewers, reviewer.ID)
	}
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, response{
		PR: responsePR{
			PullResuestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.Author.ID,
			Status:          pr.Status,
			AssignedReviews: reviewers,
		},
		ReplacedBy: newReviewer,
	})
}
