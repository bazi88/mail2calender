package calendar

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	ical "github.com/arran4/golang-ical"
)

type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Event struct {
	Summary     string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	Attendees   []string
}

type emailProcessorImpl struct {
	attachmentProcessor AttachmentProcessor
}

func NewEmailProcessor(ap AttachmentProcessor) EmailProcessor {
	return &emailProcessorImpl{
		attachmentProcessor: ap,
	}
}

func (ep *emailProcessorImpl) HandleCalendarInvite(att EmailAttachment) (*Event, error) {
	if att.ContentType != "text/calendar" {
		return nil, nil
	}

	cal, err := ical.ParseCalendar(strings.NewReader(string(att.Data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ICS file: %w", err)
	}

	for _, event := range cal.Events() {
		// Get the first event from the calendar
		startTime, _ := event.GetStartAt()
		endTime, _ := event.GetEndAt()

		attendees := make([]string, 0)
		for _, attendee := range event.Attendees() {
			attendees = append(attendees, attendee.Email())
		}

		return &Event{
			Summary:     event.GetProperty(ical.ComponentPropertySummary).Value,
			Description: event.GetProperty(ical.ComponentPropertyDescription).Value,
			StartTime:   startTime,
			EndTime:     endTime,
			Location:    event.GetProperty(ical.ComponentPropertyLocation).Value,
			Attendees:   attendees,
		}, nil
	}

	return nil, fmt.Errorf("no events found in calendar")
}

type AttachmentProcessor struct {
	storage Storage
}

func (ap *AttachmentProcessor) StartCleanupJob(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ap.deleteOldFiles(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (ap *AttachmentProcessor) deleteOldFiles(ctx context.Context) error {
	// Set retention period to 30 days
	retentionPeriod := time.Now().AddDate(0, 0, -30)

	// Get list of files older than retention period
	files, err := ap.storage.ListFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	for _, file := range files {
		// Skip files newer than retention period
		if file.CreatedAt.After(retentionPeriod) {
			continue
		}

		if err := ap.storage.Delete(ctx, file.ID); err != nil {
			// Log error but continue with other files
			log.Printf("Failed to delete old file %s: %v", file.ID, err)
			continue
		}
	}

	return nil
}

// Storage interface defines methods for file storage operations
type Storage interface {
	Save(ctx context.Context, data []byte) (string, error)
	Get(ctx context.Context, id string) ([]byte, error)
	Delete(ctx context.Context, id string) error
	ListFiles(ctx context.Context) ([]FileInfo, error)
}

// FileInfo represents metadata about a stored file
type FileInfo struct {
	ID        string
	CreatedAt time.Time
}

// EmailProcessor interface defines methods for processing email attachments
type EmailProcessor interface {
	HandleCalendarInvite(att EmailAttachment) (*Event, error)
}
