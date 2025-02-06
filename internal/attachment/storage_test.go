package attachment

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockVirusScanner struct {
	mock.Mock
}

func (m *MockVirusScanner) Scan(data []byte) (bool, error) {
	args := m.Called(data)
	return args.Bool(0), args.Error(1)
}

func TestAttachmentProcessor_QuarantineUnscannedFiles(t *testing.T) {
	// Setup
	mockStorage := &S3Storage{
		client:           &MockMinioClient{},
		bucket:           "test-bucket",
		quarantineBucket: "quarantine-bucket",
	}
	mockScanner := new(MockVirusScanner)

	// Configure mock
	mockScanner.On("Scan", mock.Anything).Return(false, errors.New("scan timeout"))

	// Create processor with mocked dependencies
	processor := NewAttachmentProcessor(mockStorage, mockScanner)

	// Test file processing
	_, err := processor.ProcessAttachment(context.Background(), []byte("test data"), ".txt")

	// Verify error contains quarantine message
	assert.ErrorContains(t, err, "quarantine")

	// Verify mock expectations
	mockScanner.AssertExpectations(t)
}
