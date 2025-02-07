package usecase

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	calendarPb "mail2calendar/internal/domain/calendar/proto"
	"mail2calendar/internal/domain/ner"
	nerClient "mail2calendar/internal/grpc/client"
)

type calendarUseCase struct {
	nerClient *nerClient.NERClient
}

func NewCalendarUseCase(nerClient *nerClient.NERClient) CalendarUseCase {
	return &calendarUseCase{
		nerClient: nerClient,
	}
}

func (u *calendarUseCase) CreateEvent(ctx context.Context, event *calendarPb.Event, userID string) (*calendarPb.Event, error) {
	if err := u.validateEvent(event); err != nil {
		return nil, err
	}

	// Extract entities from event description if provided
	if event.Description != "" {
		entities, err := u.nerClient.ExtractEntities(ctx, event.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to extract entities: %v", err)
		}

		// Update event fields based on extracted entities
		u.updateEventWithEntities(event, entities.Entities)
	}

	// Here you would typically save the event to a database
	// For now, we'll just return the event with a generated ID
	event.Id = generateEventID()
	event.Status = "confirmed"

	return event, nil
}

func (u *calendarUseCase) UpdateEvent(ctx context.Context, event *calendarPb.Event, userID string) (*calendarPb.Event, error) {
	if err := u.validateEvent(event); err != nil {
		return nil, err
	}

	if event.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "event ID is required for update")
	}

	// Here you would typically update the event in the database
	// For now, we'll just return the updated event
	return event, nil
}

func (u *calendarUseCase) DeleteEvent(ctx context.Context, eventID string, userID string) error {
	if eventID == "" {
		return status.Error(codes.InvalidArgument, "event ID is required")
	}

	// Here you would typically delete the event from the database
	return nil
}

func (u *calendarUseCase) GetEvent(ctx context.Context, eventID string, userID string) (*calendarPb.Event, error) {
	if eventID == "" {
		return nil, status.Error(codes.InvalidArgument, "event ID is required")
	}

	// Here you would typically fetch the event from the database
	// For now, we'll return a mock event
	return &calendarPb.Event{
		Id:          eventID,
		Title:       "Mock Event",
		Description: "This is a mock event",
		StartTime:   time.Now().Unix(),
		EndTime:     time.Now().Add(time.Hour).Unix(),
		Status:      "confirmed",
	}, nil
}

func (u *calendarUseCase) ListEvents(ctx context.Context, userID string, startTime int64, endTime int64, calendarID string, pageSize int32, pageToken string) ([]*calendarPb.Event, string, error) {
	if userID == "" {
		return nil, "", status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Here you would typically fetch events from the database with pagination
	// For now, we'll return an empty list
	return []*calendarPb.Event{}, "", nil
}

func (u *calendarUseCase) validateEvent(event *calendarPb.Event) error {
	if event == nil {
		return status.Error(codes.InvalidArgument, "event cannot be nil")
	}

	if event.Title == "" {
		return status.Error(codes.InvalidArgument, "event title is required")
	}

	if event.StartTime <= 0 {
		return status.Error(codes.InvalidArgument, "valid start time is required")
	}

	if event.EndTime <= 0 {
		return status.Error(codes.InvalidArgument, "valid end time is required")
	}

	if event.EndTime <= event.StartTime {
		return status.Error(codes.InvalidArgument, "end time must be after start time")
	}

	return nil
}

func (u *calendarUseCase) updateEventWithEntities(event *calendarPb.Event, entities []*ner.Entity) {
	for _, entity := range entities {
		switch entity.Label {
		case "LOCATION":
			if event.Location == "" {
				event.Location = entity.Text
			}
		case "PERSON":
			event.Attendees = append(event.Attendees, entity.Text)
		case "DATE":
			// You might want to update start/end times based on extracted dates
			// This would require additional date parsing logic
		}
	}
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
