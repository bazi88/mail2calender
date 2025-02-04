package usecase

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

// GoogleCalendarService defines the interface for Google Calendar operations
type GoogleCalendarService interface {
	// CreateCalendarEvent creates a new event in Google Calendar
	CreateCalendarEvent(ctx context.Context, event *EmailEvent, userID string) (*calendar.Event, error)

	// UpdateCalendarEvent updates an existing event
	UpdateCalendarEvent(ctx context.Context, eventID string, event *EmailEvent, userID string) (*calendar.Event, error)

	// GetCalendarEvent retrieves an event by ID
	GetCalendarEvent(ctx context.Context, eventID string, userID string) (*calendar.Event, error)

	// DeleteCalendarEvent removes an event from the calendar
	DeleteCalendarEvent(ctx context.Context, eventID string, userID string) error

	// RefreshToken refreshes the OAuth2 token
	RefreshToken(ctx context.Context, userID string) error
}

// TokenManager handles OAuth2 token operations
type TokenManager interface {
	// GetToken retrieves the stored token for a user
	GetToken(ctx context.Context, userID string) (*oauth2.Token, error)

	// SaveToken stores a new token for a user
	SaveToken(ctx context.Context, userID string, token *oauth2.Token) error

	// DeleteToken removes a user's token
	DeleteToken(ctx context.Context, userID string) error
}

type googleCalendarService struct {
	tokenManager TokenManager
	config       *oauth2.Config
}

// NewGoogleCalendarService creates a new instance of GoogleCalendarService
func NewGoogleCalendarService(tokenManager TokenManager) (GoogleCalendarService, error) {
	config, err := getOAuth2Config()
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 config: %v", err)
	}

	return &googleCalendarService{
		tokenManager: tokenManager,
		config:       config,
	}, nil
}

func (g *googleCalendarService) CreateCalendarEvent(ctx context.Context, event *EmailEvent, userID string) (*calendar.Event, error) {
	client, err := g.getClient(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %v", err)
	}

	calendarEvent := &calendar.Event{
		Summary:     event.Subject,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
		},
		Location:  event.Location,
		Attendees: make([]*calendar.EventAttendee, len(event.Attendees)),
	}

	for i, attendee := range event.Attendees {
		calendarEvent.Attendees[i] = &calendar.EventAttendee{
			Email: attendee,
		}
	}

	srv, err := calendar.New(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %v", err)
	}

	return srv.Events.Insert("primary", calendarEvent).Do()
}

func (g *googleCalendarService) UpdateCalendarEvent(ctx context.Context, eventID string, event *EmailEvent, userID string) (*calendar.Event, error) {
	// TODO: Implement update logic similar to create
	return nil, nil
}

func (g *googleCalendarService) GetCalendarEvent(ctx context.Context, eventID string, userID string) (*calendar.Event, error) {
	// TODO: Implement get logic
	return nil, nil
}

func (g *googleCalendarService) DeleteCalendarEvent(ctx context.Context, eventID string, userID string) error {
	// TODO: Implement delete logic
	return nil
}

func (g *googleCalendarService) RefreshToken(ctx context.Context, userID string) error {
	token, err := g.tokenManager.GetToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get token: %v", err)
	}

	newToken, err := g.config.TokenSource(ctx, token).Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}

	if err := g.tokenManager.SaveToken(ctx, userID, newToken); err != nil {
		return fmt.Errorf("failed to save refreshed token: %v", err)
	}

	return nil
}

func (g *googleCalendarService) getClient(ctx context.Context, userID string) (*oauth2.Client, error) {
	token, err := g.tokenManager.GetToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	return g.config.Client(ctx, token), nil
}

func getOAuth2Config() (*oauth2.Config, error) {
	// TODO: Load from secure configuration
	return &oauth2.Config{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		RedirectURL:  "your-redirect-url",
		Scopes: []string{
			calendar.CalendarScope,
			calendar.CalendarEventsScope,
		},
		Endpoint: google.Endpoint,
	}, nil
}
