package usecase

import (
	"context"
	"errors"
	"testing"

	"mail2calendar/internal/domain/ner"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNERClient is a mock implementation of client.NER
type MockNERClient struct {
	mock.Mock
}

func (m *MockNERClient) ExtractEntities(ctx context.Context, text string) (*ner.ExtractResponse, error) {
	args := m.Called(ctx, text)
	if resp, ok := args.Get(0).(*ner.ExtractResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestNew(t *testing.T) {
	mockClient := new(MockNERClient)
	useCase := New(mockClient)
	assert.NotNil(t, useCase)
}

func TestNERUseCase_ExtractEntities(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		mockResponse  *ner.ExtractResponse
		mockError     error
		expectedError bool
	}{
		{
			name: "successful extraction",
			text: "John works at Google",
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
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "client error",
			text:          "test text",
			mockResponse:  nil,
			mockError:     errors.New("client error"),
			expectedError: true,
		},
		{
			name:          "empty text",
			text:          "",
			mockResponse:  &ner.ExtractResponse{Entities: []*ner.Entity{}},
			mockError:     nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockNERClient)
			mockClient.On("ExtractEntities", mock.Anything, tt.text).Return(tt.mockResponse, tt.mockError)

			useCase := New(mockClient)
			response, err := useCase.ExtractEntities(context.Background(), tt.text)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse, response)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestNERUseCase_ExtractEntitiesFromText(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		mockResponse  *ner.ExtractResponse
		mockError     error
		expectedError bool
	}{
		{
			name: "successful extraction",
			text: "John works at Google",
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
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "client error",
			text:          "test text",
			mockResponse:  nil,
			mockError:     errors.New("client error"),
			expectedError: true,
		},
		{
			name:          "empty text",
			text:          "",
			mockResponse:  &ner.ExtractResponse{Entities: []*ner.Entity{}},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "nil response",
			text:          "test",
			mockResponse:  nil,
			mockError:     nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockNERClient)
			mockClient.On("ExtractEntities", mock.Anything, tt.text).Return(tt.mockResponse, tt.mockError)

			useCase := New(mockClient)
			entities, err := useCase.ExtractEntitiesFromText(context.Background(), tt.text)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, entities)
			} else {
				assert.NoError(t, err)
				if tt.mockResponse != nil {
					assert.Equal(t, tt.mockResponse.Entities, entities)
				} else {
					assert.Nil(t, entities)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}
