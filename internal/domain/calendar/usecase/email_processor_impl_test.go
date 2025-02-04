package usecase

import (
	"context"
	"fmt"
	"net/mail"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type mockEmailValidator struct {
	mock.Mock
}

func (m *mockEmailValidator) ValidateDKIM(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *mockEmailValidator) ValidateSPF(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *mockEmailValidator) ValidateSender(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

type mockNERService struct {
	mock.Mock
}

func (m *mockNERService) ExtractEntities(ctx context.Context, text string, language string) ([]Entity, error) {
	args := m.Called(ctx, text, language)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *mockNERService) ExtractDateTime(ctx context.Context, text string) ([]time.Time, error) {
	args := m.Called(ctx, text)
	return args.Get(0).([]time.Time), args.Error(1)
}

func (m *mockNERService) ExtractLocation(ctx context.Context, text string) (string, error) {
	args := m.Called(ctx, text)
	return args.String(0), args.Error(1)
}

func TestEmailProcessorImpl_ProcessEmail(t *testing.T) {
	tests := []struct {
		name          string
		emailContent  string
		setupMocks    func(*mockEmailValidator, *mockNERService)
		expectedEvent *EmailEvent
		expectError   bool
	}{
		{
			name: "successful processing",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Meeting at 2pm tomorrow
Content-Type: text/plain

Let's meet tomorrow at 2pm in the conference room.`,
			setupMocks: func(validator *mockEmailValidator, ner *mockNERService) {
				startTime := time.Now().Add(24 * time.Hour).Round(time.Hour).Add(14 * time.Hour)
				endTime := startTime.Add(time.Hour)

				ner.On("ExtractDateTime", mock.Anything, mock.Anything).
					Return([]time.Time{startTime, endTime}, nil)

				ner.On("ExtractLocation", mock.Anything, mock.Anything).
					Return("conference room", nil)
			},
			expectedEvent: &EmailEvent{
				Subject:   "Meeting at 2pm tomorrow",
				Location:  "conference room",
				Attendees: []string{"recipient@example.com"},
			},
			expectError: false,
		},
		{
			name:          "invalid email format",
			emailContent:  "invalid email content",
			setupMocks:    func(validator *mockEmailValidator, ner *mockNERService) {},
			expectedEvent: nil,
			expectError:   true,
		},
		{
			name: "NER service error",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Meeting tomorrow
Content-Type: text/plain

Meeting details.`,
			setupMocks: func(validator *mockEmailValidator, ner *mockNERService) {
				ner.On("ExtractDateTime", mock.Anything, mock.Anything).
					Return([]time.Time{}, fmt.Errorf("NER service error"))
			},
			expectedEvent: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			validator := new(mockEmailValidator)
			ner := new(mockNERService)
			tt.setupMocks(validator, ner)

			// Create processor
			processor := NewEmailProcessorImpl(validator, ner)

			// Process email
			event, err := processor.ProcessEmail(context.Background(), tt.emailContent)

			// Check results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, event)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, event)
				if tt.expectedEvent != nil {
					assert.Equal(t, tt.expectedEvent.Subject, event.Subject)
					assert.Equal(t, tt.expectedEvent.Location, event.Location)
					assert.ElementsMatch(t, tt.expectedEvent.Attendees, event.Attendees)
				}
			}

			// Verify mock expectations
			validator.AssertExpectations(t)
			ner.AssertExpectations(t)
		})
	}
}

func TestEmailProcessorImpl_ValidateEmail(t *testing.T) {
	tests := []struct {
		name         string
		emailContent string
		setupMocks   func(*mockEmailValidator)
		expectError  bool
	}{
		{
			name:         "successful validation",
			emailContent: "valid email content",
			setupMocks: func(validator *mockEmailValidator) {
				validator.On("ValidateDKIM", mock.Anything).Return(nil)
				validator.On("ValidateSPF", mock.Anything).Return(nil)
				validator.On("ValidateSender", mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name:         "DKIM validation fails",
			emailContent: "invalid email content",
			setupMocks: func(validator *mockEmailValidator) {
				validator.On("ValidateDKIM", mock.Anything).Return(fmt.Errorf("DKIM error"))
			},
			expectError: true,
		},
		{
			name:         "SPF validation fails",
			emailContent: "invalid email content",
			setupMocks: func(validator *mockEmailValidator) {
				validator.On("ValidateDKIM", mock.Anything).Return(nil)
				validator.On("ValidateSPF", mock.Anything).Return(fmt.Errorf("SPF error"))
			},
			expectError: true,
		},
		{
			name:         "sender validation fails",
			emailContent: "invalid email content",
			setupMocks: func(validator *mockEmailValidator) {
				validator.On("ValidateDKIM", mock.Anything).Return(nil)
				validator.On("ValidateSPF", mock.Anything).Return(nil)
				validator.On("ValidateSender", mock.Anything).Return(fmt.Errorf("sender error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			validator := new(mockEmailValidator)
			tt.setupMocks(validator)

			// Create processor
			processor := NewEmailProcessorImpl(validator, new(mockNERService))

			// Validate email
			err := processor.ValidateEmail(context.Background(), tt.emailContent)

			// Check result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			validator.AssertExpectations(t)
		})
	}
}

func TestEmailProcessorImpl_extractEmailContent(t *testing.T) {
	tests := []struct {
		name         string
		emailContent string
		expectedText string
		expectedHTML string
		hasAttach    bool
	}{
		{
			name: "plain text email",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Test Email
Content-Type: text/plain

This is a test email.`,
			expectedText: "This is a test email.",
			expectedHTML: "",
			hasAttach:    false,
		},
		{
			name: "HTML email",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Test Email
Content-Type: text/html

<html><body>This is a test email.</body></html>`,
			expectedText: "",
			expectedHTML: "<html><body>This is a test email.</body></html>",
			hasAttach:    false,
		},
		{
			name: "multipart email",
			emailContent: `From: sender@example.com
To: recipient@example.com
Subject: Test Email
Content-Type: multipart/mixed; boundary="boundary"

--boundary
Content-Type: text/plain

Plain text content
--boundary
Content-Type: text/html

<html><body>HTML content</body></html>
--boundary--`,
			expectedText: "Plain text content",
			expectedHTML: "<html><body>HTML content</body></html>",
			hasAttach:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create concrete implementation
			processor := &emailProcessorImpl{
				validator:  new(mockEmailValidator),
				nerService: new(mockNERService),
			}

			msg, err := processor.parseEmail(context.Background(), tt.emailContent)
			assert.NoError(t, err)

			content, err := processor.extractEmailContent(context.Background(), msg)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedText, content.PlainText)
			assert.Equal(t, tt.expectedHTML, content.HTML)
			assert.Equal(t, tt.hasAttach, len(content.Attachments) > 0)
		})
	}
}

func TestEmailProcessorImpl_extractAttendees(t *testing.T) {
	tests := []struct {
		name           string
		to             string
		cc             string
		expectedEmails []string
	}{
		{
			name:           "single recipient",
			to:             "user@example.com",
			cc:             "",
			expectedEmails: []string{"user@example.com"},
		},
		{
			name: "multiple recipients",
			to:   "user1@example.com, user2@example.com",
			cc:   "user3@example.com",
			expectedEmails: []string{
				"user1@example.com",
				"user2@example.com",
				"user3@example.com",
			},
		},
		{
			name:           "no recipients",
			to:             "",
			cc:             "",
			expectedEmails: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create concrete implementation
			processor := &emailProcessorImpl{
				validator:  new(mockEmailValidator),
				nerService: new(mockNERService),
			}

			header := make(mail.Header)
			if tt.to != "" {
				header["To"] = []string{tt.to}
			}
			if tt.cc != "" {
				header["Cc"] = []string{tt.cc}
			}

			attendees := processor.extractAttendees(header)
			assert.ElementsMatch(t, tt.expectedEmails, attendees)
		})
	}
}
