package handler

import (
	"mono-golang/internal/domain/ner"
	"mono-golang/internal/middleware"

	"github.com/go-chi/chi/v5"
)

// Register registers the NER routes
func Register(r chi.Router, useCase ner.UseCase, rateLimiter middleware.RateLimiter) {
	handler := New(useCase)

	r.Route("/api/v1/ner", func(r chi.Router) {
		r.Use(rateLimiter.Limit)
		r.Post("/extract", handler.Extract)
	})
}
