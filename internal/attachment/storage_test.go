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

func TestAttachmentProcessor_ProcessAttachment(t *testing.T) {
	// Setup
	mockStorage := new(mockStorage)
	mockScanner := new(MockVirusScanner)
	processor := NewAttachmentProcessor(mockStorage, mockScanner)

	testData := []byte("test data")
	testExt := ".txt"
	testFileID := "test-file-id"

	// Test cases
	tests := []struct {
		name        string
		scanResult  bool
		scanError   error
		saveError   error
		expectError bool
		expectID    string
	}{
		{
			name:        "successful scan and save",
			scanResult:  true,
			scanError:   nil,
			saveError:   nil,
			expectError: false,
			expectID:    testFileID,
		},
		{
			name:        "virus detected",
			scanResult:  false,
			scanError:   nil,
			expectError: true,
			expectID:    "",
		},
		{
			name:        "scan error",
			scanResult:  false,
			scanError:   errors.New("scan failed"),
			expectError: true,
			expectID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure mocks
			mockScanner.On("Scan", testData).Return(tt.scanResult, tt.scanError).Once()
			if tt.scanResult && tt.scanError == nil {
				mockStorage.On("Save", mock.Anything, testData, testExt).Return(testFileID, tt.saveError).Once()
			}

			// Call the method
			fileID, err := processor.ProcessAttachment(context.Background(), testData, testExt)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, fileID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectID, fileID)
			}

			// Verify mock expectations
			mockScanner.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}

// mockStorage is a mock implementation of Storage interface
type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Save(ctx context.Context, data []byte, ext string) (string, error) {
	args := m.Called(ctx, data, ext)
	return args.String(0), args.Error(1)
}

func (m *mockStorage) Get(ctx context.Context, id string) ([]byte, string, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *mockStorage) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockStorage) ListFiles(ctx context.Context) ([]FileInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]FileInfo), args.Error(1)
}
