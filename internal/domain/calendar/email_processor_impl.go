package calendar

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	ical "github.com/arran4/golang-ical"
)

// EmailAttachment đại diện cho một tệp đính kèm email
type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Event đại diện cho một sự kiện lịch
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
	logger              *log.Logger
}

// NewEmailProcessor tạo một email processor mới
func NewEmailProcessor(ap AttachmentProcessor) EmailProcessor {
	return &emailProcessorImpl{
		attachmentProcessor: ap,
		logger:              log.New(log.Writer(), "[EmailProcessor] ", log.LstdFlags),
	}
}

func (ep *emailProcessorImpl) HandleCalendarInvite(att EmailAttachment) (*Event, error) {
	if att.ContentType != "text/calendar" {
		ep.logger.Printf("Không phải file calendar: %s", att.ContentType)
		return nil, nil
	}

	cal, err := ical.ParseCalendar(strings.NewReader(string(att.Data)))
	if err != nil {
		ep.logger.Printf("Lỗi khi parse file ICS: %v", err)
		return nil, fmt.Errorf("failed to parse ICS file: %w", err)
	}

	events := cal.Events()
	if len(events) == 0 {
		ep.logger.Print("Không tìm thấy sự kiện trong calendar")
		return nil, fmt.Errorf("no events found in calendar")
	}

	event := events[0]
	startTime, err := event.GetStartAt()
	if err != nil {
		ep.logger.Printf("Lỗi khi lấy thời gian bắt đầu: %v", err)
		return nil, fmt.Errorf("failed to get start time: %w", err)
	}

	endTime, err := event.GetEndAt()
	if err != nil {
		ep.logger.Printf("Lỗi khi lấy thời gian kết thúc: %v", err)
		return nil, fmt.Errorf("failed to get end time: %w", err)
	}

	attendees := make([]string, 0)
	for _, attendee := range event.Attendees() {
		if email := attendee.Email(); email != "" {
			attendees = append(attendees, email)
		}
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

// AttachmentProcessor xử lý các tệp đính kèm
type AttachmentProcessor struct {
	storage Storage
	logger  *log.Logger
}

func (ap *AttachmentProcessor) StartCleanupJob(ctx context.Context) {
	ap.logger = log.New(log.Writer(), "[AttachmentProcessor] ", log.LstdFlags)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := ap.deleteOldFiles(ctx); err != nil {
					ap.logger.Printf("Lỗi khi xóa file cũ: %v", err)
				}
			case <-ctx.Done():
				ap.logger.Print("Cleanup job stopped")
				return
			}
		}
	}()
}

func (ap *AttachmentProcessor) deleteOldFiles(ctx context.Context) error {
	// Set retention period to 30 days
	retentionPeriod := time.Now().AddDate(0, 0, -30)

	files, err := ap.storage.ListFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	var deleteErrors []error
	for _, file := range files {
		if file.CreatedAt.After(retentionPeriod) {
			continue
		}

		if err := ap.storage.Delete(ctx, file.ID); err != nil {
			deleteErrors = append(deleteErrors, fmt.Errorf("failed to delete file %s: %w", file.ID, err))
			ap.logger.Printf("Lỗi khi xóa file %s: %v", file.ID, err)
		}
	}

	if len(deleteErrors) > 0 {
		return fmt.Errorf("multiple errors occurred while deleting files: %v", deleteErrors)
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
