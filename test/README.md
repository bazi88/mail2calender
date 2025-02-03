# /test

Thư mục này chứa các công cụ và tài nguyên testing.

## Cấu Trúc

```
/mock
  email_mock.go     # Email service mocks
  calendar_mock.go  # Calendar service mocks
  ner_mock.go       # NER service mocks

/fixtures
  users.json        # Test user data
  emails.json       # Sample emails
  events.json       # Calendar events

/integration
  email_test.go     # Email flow tests  
  calendar_test.go  # Calendar sync tests
  api_test.go       # API integration tests

/e2e
  scenarios/        # Test scenarios
  helpers/          # Test utilities
  reports/          # Test reports
```

## Loại Test

### Unit Tests
- Domain logic
- Service implementations
- Utility functions
- Error handling

### Integration Tests
- API endpoints
- Database operations
- External services
- Background jobs

### E2E Tests
- Critical flows
- User scenarios
- Performance tests
- Error scenarios

## Test Data

### Fixtures
- Sample data
- Test cases
- Expected results
- Error cases

### Mocks
- Service mocks
- API responses
- Error conditions
- Test doubles

## Sử Dụng

```bash
# Run all tests
go test ./...

# Run specific test
go test ./test/integration/email_test.go

# Run with coverage
go test -cover ./...

# Generate mock
mockgen -source=internal/domain/email.go
``` 