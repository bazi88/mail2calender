package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "mail2calendar/internal/domain/calendar/proto"
	"mail2calendar/internal/domain/calendar/usecase"
)

type CalendarHandler struct {
	pb.UnimplementedCalendarServiceServer
	useCase usecase.CalendarUseCase
}

func NewCalendarHandler(useCase usecase.CalendarUseCase) *CalendarHandler {
	return &CalendarHandler{
		useCase: useCase,
	}
}

func (h *CalendarHandler) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event cannot be nil")
	}

	event, err := h.useCase.CreateEvent(ctx, req.Event, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateEventResponse{
		Event: event,
	}, nil
}

func (h *CalendarHandler) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	if req.Event == nil {
		return nil, status.Error(codes.InvalidArgument, "event cannot be nil")
	}

	event, err := h.useCase.UpdateEvent(ctx, req.Event, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateEventResponse{
		Event: event,
	}, nil
}

func (h *CalendarHandler) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event ID cannot be empty")
	}

	err := h.useCase.DeleteEvent(ctx, req.EventId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteEventResponse{
		Success: true,
	}, nil
}

func (h *CalendarHandler) GetEvent(ctx context.Context, req *pb.GetEventRequest) (*pb.GetEventResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event ID cannot be empty")
	}

	event, err := h.useCase.GetEvent(ctx, req.EventId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetEventResponse{
		Event: event,
	}, nil
}

func (h *CalendarHandler) ListEvents(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID cannot be empty")
	}

	events, nextPageToken, err := h.useCase.ListEvents(ctx, req.UserId, req.StartTime, req.EndTime, req.CalendarId, req.PageSize, req.PageToken)
	if err != nil {
		return nil, err
	}

	return &pb.ListEventsResponse{
		Events:        events,
		NextPageToken: nextPageToken,
	}, nil
}
