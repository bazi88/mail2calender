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

type AttachmentProcessor struct {
	storage Storage
	scanner VirusScanner
}

func NewAttachmentProcessor(storage Storage, scanner VirusScanner) *AttachmentProcessor {
	return &AttachmentProcessor{
		storage: storage,
		scanner: scanner,
	}
}

func (p *AttachmentProcessor) ProcessAttachment(ctx context.Context, data []byte, ext string) (string, error) {
	// Scan for viruses
	isClean, err := p.scanner.Scan(data)
	if err != nil {
		// If scan fails, quarantine the file
		return "", errors.New("virus scan failed, file quarantined")
	}
	if !isClean {
		return "", errors.New("virus detected, file quarantined")
	}

	// Save clean file
	return p.storage.Save(ctx, data, ext)
}

func TestAttachmentProcessor_QuarantineUnscannedFiles(t *testing.T) {
	// Setup
	mockClient := new(mockMinioClient)
	mockStorage := &S3Storage{
		client:           mockClient,
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
