package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	calendarPb "mono-golang/internal/domain/calendar/proto"
	"mono-golang/internal/domain/calendar/usecase"
)

type CalendarHandler struct {
	calendarPb.UnimplementedCalendarServiceServer
	useCase usecase.CalendarUseCase
}

func NewCalendarHandler(useCase usecase.CalendarUseCase) *CalendarHandler {
	return &CalendarHandler{
		useCase: useCase,
	}
}

func (h *CalendarHandler) CreateEvent(ctx context.Context, req *calendarPb.CreateEventRequest) (*calendarPb.CreateEventResponse, error) {
	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event cannot be nil")
	}

	event, err := h.useCase.CreateEvent(ctx, req.Event, req.UserId)
	if err != nil {
		return nil, err
	}

	return &calendarPb.CreateEventResponse{
		Event: event,
	}, nil
}

func (h *CalendarHandler) UpdateEvent(ctx context.Context, req *calendarPb.UpdateEventRequest) (*calendarPb.UpdateEventResponse, error) {
	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event cannot be nil")
	}

	event, err := h.useCase.UpdateEvent(ctx, req.Event, req.UserId)
	if err != nil {
		return nil, err
	}

	return &calendarPb.UpdateEventResponse{
		Event: event,
	}, nil
}

func (h *CalendarHandler) DeleteEvent(ctx context.Context, req *calendarPb.DeleteEventRequest) (*calendarPb.DeleteEventResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event ID cannot be empty")
	}

	err := h.useCase.DeleteEvent(ctx, req.EventId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &calendarPb.DeleteEventResponse{
		Success: true,
	}, nil
}

func (h *CalendarHandler) GetEvent(ctx context.Context, req *calendarPb.GetEventRequest) (*calendarPb.GetEventResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event ID cannot be empty")
	}

	event, err := h.useCase.GetEvent(ctx, req.EventId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &calendarPb.GetEventResponse{
		Event: event,
	}, nil
}

func (h *CalendarHandler) ListEvents(ctx context.Context, req *calendarPb.ListEventsRequest) (*calendarPb.ListEventsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID cannot be empty")
	}

	events, nextPageToken, err := h.useCase.ListEvents(ctx, req.UserId, req.StartTime, req.EndTime, req.CalendarId, req.PageSize, req.PageToken)
	if err != nil {
		return nil, err
	}

	return &calendarPb.ListEventsResponse{
		Events:        events,
		NextPageToken: nextPageToken,
	}, nil
}
