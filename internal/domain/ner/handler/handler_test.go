package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mail2calendar/internal/domain/ner"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockNERUseCase struct {
	mock.Mock
}

func (m *MockNERUseCase) ExtractEntities(ctx context.Context, text string) (*ner.ExtractResponse, error) {
	args := m.Called(ctx, text)
	if resp := args.Get(0); resp != nil {
		return resp.(*ner.ExtractResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestHandler_ExtractEntities(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockNERUseCase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful extraction",
			requestBody: extractRequest{
				Text: "test text",
			},
			setupMock: func(m *MockNERUseCase) {
				m.On("ExtractEntities", mock.Anything, "test text").Return(&ner.ExtractResponse{
					Entities: []*ner.Entity{
						{
							Text:  "test",
							Label: "TEST",
							Start: 0,
							End:   4,
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"entities":[{"text":"test","label":"TEST","start":0,"end":4}]}`,
		},
		{
			name: "invalid request body",
			requestBody: struct {
				InvalidField string `json:"invalid_field"`
			}{
				InvalidField: "test",
			},
			setupMock:      func(m *MockNERUseCase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockNERUseCase)
			tt.setupMock(mockUseCase)

			handler := &Handler{
				useCase: mockUseCase,
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/ner/extract", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.ExtractEntities(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
			mockUseCase.AssertExpectations(t)
		})
	}
}
