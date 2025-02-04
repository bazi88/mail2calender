package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mono-golang/internal/domain/ner"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNERUseCase is a mock implementation of ner.UseCase
type MockNERUseCase struct {
	mock.Mock
}

func (m *MockNERUseCase) ExtractEntities(ctx context.Context, text string) (*ner.ExtractResponse, error) {
	args := m.Called(ctx, text)
	if resp, ok := args.Get(0).(*ner.ExtractResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNERUseCase) ExtractEntitiesFromText(ctx context.Context, text string) ([]*ner.Entity, error) {
	args := m.Called(ctx, text)
	if entities, ok := args.Get(0).([]*ner.Entity); ok {
		return entities, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestNew(t *testing.T) {
	mockUseCase := new(MockNERUseCase)
	handler := New(mockUseCase)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.validate)
	assert.Equal(t, mockUseCase, handler.useCase)
}

func TestHandler_Extract(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      bool
		mockResponse   *ner.ExtractResponse
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful extraction",
			requestBody: ExtractRequest{
				Text: "John works at Google",
			},
			setupMock: true,
			mockResponse: &ner.ExtractResponse{
				Entities: []*ner.Entity{
					{
						Text:  "John",
						Label: "PERSON",
						Start: 0,
						End:   4,
					},
					{
						Text:  "Google",
						Label: "ORG",
						Start: 11,
						End:   17,
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: &ner.ExtractResponse{
				Entities: []*ner.Entity{
					{
						Text:  "John",
						Label: "PERSON",
						Start: 0,
						End:   4,
					},
					{
						Text:  "Google",
						Label: "ORG",
						Start: 11,
						End:   17,
					},
				},
			},
		},
		{
			name: "empty text",
			requestBody: ExtractRequest{
				Text: "",
			},
			setupMock:      false,
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			setupMock:      false,
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "usecase error",
			requestBody: ExtractRequest{
				Text: "test text",
			},
			setupMock:      true,
			mockResponse:   nil,
			mockError:      errors.New("usecase error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockNERUseCase)
			handler := New(mockUseCase)

			// Prepare request
			var body []byte
			var err error
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/ner/extract", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Set up mock expectations
			if tt.setupMock {
				if reqBody, ok := tt.requestBody.(ExtractRequest); ok {
					mockUseCase.On("ExtractEntities", mock.Anything, reqBody.Text).Return(tt.mockResponse, tt.mockError)
				}
			}

			// Execute request
			handler.Extract(rec, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != nil {
				var response ner.ExtractResponse
				err := json.NewDecoder(rec.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, &response)
			}

			mockUseCase.AssertExpectations(t)
		})
	}
}
