package router

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/internal/transport"
	"github.com/sudo-odner/Backend-trainee-assignment-avito-2025/pkg/logger/sl"
)

func (router *Router) StatGetReviews(w http.ResponseWriter, r *http.Request) {
	type responseReviewer struct {
		UserID      string `json:"user_id"`
		ReviewCount int    `json:"review_count"`
	}
	type response struct {
		ReviewStat []responseReviewer `json:"review_stat"`
	}
	stat, err := router.storage.GetReviewStat()
	if err != nil {
		router.log.Error("Failed get reviewStat", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, transport.ErrResponse{
			Code:    transport.SERVER_ERROR,
			Message: "failed get review stat",
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	reviewersStat := make([]responseReviewer, 0, len(stat))
	for _, reviewer := range stat {
		reviewersStat = append(reviewersStat, responseReviewer{
			UserID:      reviewer.UserID,
			ReviewCount: reviewer.ReviewCount,
		})
	}
	render.JSON(w, r, response{
		ReviewStat: reviewersStat,
	})
}
