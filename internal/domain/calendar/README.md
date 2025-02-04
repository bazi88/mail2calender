# Calendar Domain

This package implements email to calendar event processing with advanced features like NER (Named Entity Recognition), conflict detection, and comprehensive error handling.

## Core Components

### Email Processing

- Full MIME support (plain text, HTML, rich text)
- Attachment and metadata extraction
- HTML link parsing
- Email validation (DKIM, SPF)

### Named Entity Recognition (NER)

- Bilingual support (Vietnamese and English)
- BERT model integration for Vietnamese
- Entity extraction with confidence scores
- Time and location recognition

### Time Processing

- Multiple timezone support
- Various time format parsing
- Recurring events handling
- Working hours management

### Conflict Detection

- Calendar event conflict checking
- Alternative time suggestions
- All-day event handling
- Working hours consideration

### Error Handling

- Domain-specific error types
- Structured error context
- Retry mechanisms
- Detailed error logging

## Usage Example

```go
// Create dependencies
validator := NewEmailValidator()
nerService := NewNERService("http://localhost:8000")
calendarService := NewCalendarService(googleCalendar)
logger := NewLogger(tracer)

// Create email processor
processor := NewEmailProcessorImpl(validator, nerService)

// Process email
event, err := processor.ProcessEmail(ctx, emailContent)
if err != nil {
    logger.ErrorWithContext(ctx, "Failed to process email", err)
    return err
}

// Check for conflicts
checker := NewConflictChecker(calendarService)
result, err := checker.CheckConflicts(ctx, event)
if err != nil {
    logger.ErrorWithContext(ctx, "Failed to check conflicts", err)
    return err
}

if result.HasConflict {
    logger.Info("Conflict detected", Fields{
        "event_id": event.ID,
        "conflict_with": result.ConflictingEvent.ID,
        "alternatives": len(result.Alternatives),
    })
    // Handle conflict...
}
```

## Error Handling Example

```go
if err := validateEmail(email); err != nil {
    return errors.NewInvalidEmailError("Invalid email format").
        WithDetails(map[string]interface{}{
            "email": email,
            "validation_error": err.Error(),
        })
}
```

## Testing

The package includes comprehensive test coverage:

- Unit tests for all components
- Integration tests for NER service
- Error handling tests
- Logging tests

Run tests:

```bash
go test ./...
```

## Configuration

### NER Service

- `NER_SERVICE_URL`: URL of the NER service
- `NER_TIMEOUT`: Request timeout in seconds

### Google Calendar

- `GOOGLE_CALENDAR_CLIENT_ID`: OAuth client ID
- `GOOGLE_CALENDAR_CLIENT_SECRET`: OAuth client secret
- `GOOGLE_CALENDAR_REDIRECT_URL`: OAuth redirect URL

### Logging

- `LOG_LEVEL`: Minimum log level (debug, info, warn, error)
- `LOG_FORMAT`: Log format (json, console)

## Dependencies

- go.uber.org/zap: Structured logging
- go.opentelemetry.io/otel: Tracing and metrics
- golang.org/x/oauth2: Google Calendar authentication
- google.golang.org/api: Google Calendar API

## Future Improvements

- Add caching for frequently accessed calendar data
- Implement batch processing for multiple emails
- Add support for more calendar providers
- Enhance NER accuracy with custom training data
- Add performance benchmarking and optimization
