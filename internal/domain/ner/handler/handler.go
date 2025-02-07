package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"mail2calendar/internal/delivery/http/middleware"
	"mail2calendar/internal/domain/ner"

	"github.com/go-chi/chi/v5"
)

type NERUseCase interface {
	ExtractEntities(ctx context.Context, text string) (*ner.ExtractResponse, error)
}

type Handler struct {
	useCase NERUseCase
}

func RegisterRoutes(r chi.Router, uc NERUseCase, rateLimiter *middleware.RedisRateLimiter) {
	handler := &Handler{
		useCase: uc,
	}

	r.Route("/api/v1/ner", func(r chi.Router) {
		r.Use(rateLimiter.Limit)
		r.Post("/extract", handler.ExtractEntities)
	})
}

type extractRequest struct {
	Text string `json:"text"`
}

func (h *Handler) ExtractEntities(w http.ResponseWriter, r *http.Request) {
	var req extractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entities, err := h.useCase.ExtractEntities(r.Context(), req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
