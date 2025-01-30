package handler

import (
	"encoding/json"
	"mono-golang/internal/domain/ner"
	"mono-golang/internal/utility/respond"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ExtractRequest struct {
	Text string `json:"text" validate:"required"`
}

type Handler struct {
	useCase  ner.UseCase
	validate *validator.Validate
}

// New creates a new NER handler
func New(useCase ner.UseCase) *Handler {
	return &Handler{
		useCase:  useCase,
		validate: validator.New(),
	}
}

// Extract handles the entity extraction request
func (h *Handler) Extract(w http.ResponseWriter, r *http.Request) {
	var req ExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respond.WriteError(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.useCase.ExtractEntities(r.Context(), req.Text)
	if err != nil {
		respond.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	respond.JSON(w, http.StatusOK, response)
}
