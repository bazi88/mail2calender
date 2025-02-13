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

type Entity struct {
	Text  string `json:"Text"`
	Label string `json:"Label"`
	Start int    `json:"Start"`
	End   int    `json:"End"`
}

type MockNERUseCase struct {
	mock.Mock
}

func (m *MockNERUseCase) ExtractEntities(ctx context.Context, text string) (*ner.ExtractResponse, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*ner.ExtractResponse), args.Error(1)
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
				Text: "test",
			},
			setupMock: func(m *MockNERUseCase) {
				m.On("ExtractEntities", mock.Anything, "test").Return(
					&ner.ExtractResponse{
						Entities: []*ner.Entity{
							{
								Text:  "test",
								Label: "TEST",
								Start: 0,
								End:   4,
							},
						},
					},
					nil,
				)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Entities":[{"Text":"test","Label":"TEST","Start":0,"End":4}]}` + "\n",
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			setupMock:      func(m *MockNERUseCase) {}, // No need to setup mock as ExtractEntities won't be called
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body\n",
		},
		{
			name: "empty text field",
			requestBody: extractRequest{
				Text: "",
			},
			setupMock:      func(m *MockNERUseCase) {}, // No need to setup mock as ExtractEntities won't be called
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Text field is required\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockNERUseCase)
			tt.setupMock(mockUseCase)
			handler := &Handler{
				useCase: mockUseCase,
			}

			var body []byte
			var err error
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/extract", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()

			handler.ExtractEntities(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
			mockUseCase.AssertExpectations(t)
		})
	}
}
