package proto

import (
	"time"
)

type CalendarEvent struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Location    string    `json:"location"`
}

type NewCreateEventRequest struct {
	Event *CalendarEvent `json:"event"`
}

type CreateEventResponseV2 struct {
	EventID string `json:"event_id"`
}

type GetEventRequestV2 struct {
	EventID string `json:"event_id"`
}

type GetEventResponseV2 struct {
	Event *Event `json:"event"`
}
