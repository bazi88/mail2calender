package handler

import (
	"encoding/json"
	"net/http"

	"mail2calendar/internal/domain/calendar/proto"
	"mail2calendar/internal/domain/calendar/service"
)

// HTTPCalendarHandler xử lý các yêu cầu HTTP cho calendar service
type HTTPCalendarHandler struct {
	svc service.CalendarService
}

// NewHTTPCalendarHandler tạo một HTTPCalendarHandler mới
func NewHTTPCalendarHandler(svc service.CalendarService) *HTTPCalendarHandler {
	return &HTTPCalendarHandler{
		svc: svc,
	}
}

// CreateEvent xử lý yêu cầu tạo event mới
func (h *HTTPCalendarHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
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

// GetEvent xử lý yêu cầu lấy thông tin event
func (h *HTTPCalendarHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
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
