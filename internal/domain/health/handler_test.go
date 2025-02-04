package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUseCase is a mock implementation of UseCase
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Readiness() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewHandler(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewHandler(mockUseCase)
	assert.NotNil(t, handler)
	assert.Equal(t, mockUseCase, handler.useCase)
}

func TestHandler_Health(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewHandler(mockUseCase)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]int
	err := json.NewDecoder(rec.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 200, response["status"])
}

func TestHandler_Readiness(t *testing.T) {
	tests := []struct {
		name           string
		mockError      error
		expectedStatus int
		expectedBody   map[string]int
	}{
		{
			name:           "successful readiness check",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]int{"status": 200},
		},
		{
			name:           "database error",
			mockError:      errors.New("database connection error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockUseCase)
			mockUseCase.On("Readiness").Return(tt.mockError)

			handler := NewHandler(mockUseCase)

			req := httptest.NewRequest(http.MethodGet, "/api/health/readiness", nil)
			rec := httptest.NewRecorder()

			handler.Readiness(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != nil {
				var response map[string]int
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}

			mockUseCase.AssertExpectations(t)
		})
	}
}
