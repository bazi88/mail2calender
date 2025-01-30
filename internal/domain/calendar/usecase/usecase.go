package usecase

import (
	"context"

	calendarPb "mono-golang/internal/domain/calendar/proto"
)

// CalendarUseCase defines the interface for calendar operations
type CalendarUseCase interface {
	CreateEvent(ctx context.Context, event *calendarPb.Event, userID string) (*calendarPb.Event, error)
	UpdateEvent(ctx context.Context, event *calendarPb.Event, userID string) (*calendarPb.Event, error)
	DeleteEvent(ctx context.Context, eventID string, userID string) error
	GetEvent(ctx context.Context, eventID string, userID string) (*calendarPb.Event, error)
	ListEvents(ctx context.Context, userID string, startTime int64, endTime int64, calendarID string, pageSize int32, pageToken string) ([]*calendarPb.Event, string, error)
}
