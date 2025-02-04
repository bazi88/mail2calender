package health

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Readiness() error {
	args := m.Called()
	return args.Error(0)
}

func TestNew(t *testing.T) {
	mockRepo := new(MockRepository)
	useCase := New(mockRepo)
	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.healthRepo)
}

func TestHealth_Readiness(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		expectedError error
	}{
		{
			name:          "successful readiness check",
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "database error",
			mockError:     errors.New("database connection error"),
			expectedError: errors.New("database connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockRepo.On("Readiness").Return(tt.mockError)

			useCase := New(mockRepo)
			err := useCase.Readiness()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
