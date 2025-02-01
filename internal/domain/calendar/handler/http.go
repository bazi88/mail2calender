package handler

import (
	"encoding/json"
	"net/http"

	"mono-golang/internal/domain/calendar/proto"
	"mono-golang/internal/domain/calendar/service"
)

type HttpCalendarHandler struct {
	svc service.CalendarService
}

func NewHttpCalendarHandler(svc service.CalendarService) *HttpCalendarHandler {
	return &HttpCalendarHandler{svc: svc}
}

func (h *HttpCalendarHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req proto.NewCreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.svc.CreateEvent(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *HttpCalendarHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("event_id")

	resp, err := h.svc.GetEvent(r.Context(), &proto.GetEventRequestV2{EventID: eventID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
