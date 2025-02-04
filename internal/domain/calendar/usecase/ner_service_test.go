package usecase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNERService_ExtractEntities(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		language       string
		mockResponse   nerResponse
		expectedError  bool
		expectedResult []Entity
	}{
		{
			name:     "successful extraction",
			text:     "Let's meet at Starbucks tomorrow at 2pm",
			language: "vi",
			mockResponse: nerResponse{
				Entities: []Entity{
					{
						Text:       "Starbucks",
						Label:      "LOC",
						Start:      14,
						End:        23,
						Confidence: 0.95,
					},
					{
						Text:       "tomorrow at 2pm",
						Label:      "TIME",
						Start:      24,
						End:        38,
						Confidence: 0.90,
					},
				},
				ProcessingTime: 0.05,
			},
			expectedError: false,
			expectedResult: []Entity{
				{
					Text:       "Starbucks",
					Label:      "LOC",
					Start:      14,
					End:        23,
					Confidence: 0.95,
				},
				{
					Text:       "tomorrow at 2pm",
					Label:      "TIME",
					Start:      24,
					End:        38,
					Confidence: 0.90,
				},
			},
		},
		{
			name:           "server error",
			text:           "test text",
			language:       "vi",
			mockResponse:   nerResponse{},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedError {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Check request
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/v1/extract", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Return mock response
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create service with mock server URL
			service := NewNERService(server.URL)

			// Test extraction
			entities, err := service.ExtractEntities(context.Background(), tt.text, tt.language)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, entities)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, entities)
			}
		})
	}
}

func TestNERService_ExtractDateTime(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name          string
		text          string
		mockResponse  nerResponse
		expectedError bool
		expectedDates []time.Time
	}{
		{
			name: "extract date and time",
			text: "Meeting tomorrow at 2pm",
			mockResponse: nerResponse{
				Entities: []Entity{
					{
						Text:       "tomorrow at 2pm",
						Label:      "TIME",
						Start:      8,
						End:        22,
						Confidence: 0.95,
					},
				},
			},
			expectedError: false,
			expectedDates: []time.Time{
				time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 14, 0, 0, 0, time.Local),
			},
		},
		{
			name: "multiple dates",
			text: "Meeting on Monday at 10am, ends at 11am",
			mockResponse: nerResponse{
				Entities: []Entity{
					{
						Text:       "Monday at 10am",
						Label:      "TIME",
						Start:      11,
						End:        24,
						Confidence: 0.95,
					},
					{
						Text:       "11am",
						Label:      "TIME",
						Start:      34,
						End:        38,
						Confidence: 0.95,
					},
				},
			},
			expectedError: false,
			expectedDates: []time.Time{
				time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.Local),
				time.Date(now.Year(), now.Month(), now.Day(), 11, 0, 0, 0, time.Local),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedError {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			service := NewNERService(server.URL)
			dates, err := service.ExtractDateTime(context.Background(), tt.text)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, dates)
			} else {
				assert.NoError(t, err)
				assert.Len(t, dates, len(tt.expectedDates))
				for i, expectedDate := range tt.expectedDates {
					assert.Equal(t, expectedDate.Hour(), dates[i].Hour())
					assert.Equal(t, expectedDate.Minute(), dates[i].Minute())
				}
			}
		})
	}
}

func TestNERService_ExtractLocation(t *testing.T) {
	tests := []struct {
		name             string
		text             string
		mockResponse     nerResponse
		expectedError    bool
		expectedLocation string
	}{
		{
			name: "simple location",
			text: "Meeting at Starbucks",
			mockResponse: nerResponse{
				Entities: []Entity{
					{
						Text:       "Starbucks",
						Label:      "LOC",
						Start:      11,
						End:        20,
						Confidence: 0.95,
					},
				},
			},
			expectedError:    false,
			expectedLocation: "Starbucks",
		},
		{
			name: "multiple locations - highest confidence",
			text: "Meeting at Starbucks or Coffee Bean",
			mockResponse: nerResponse{
				Entities: []Entity{
					{
						Text:       "Starbucks",
						Label:      "LOC",
						Start:      11,
						End:        20,
						Confidence: 0.95,
					},
					{
						Text:       "Coffee Bean",
						Label:      "LOC",
						Start:      24,
						End:        35,
						Confidence: 0.90,
					},
				},
			},
			expectedError:    false,
			expectedLocation: "Starbucks",
		},
		{
			name: "no location found",
			text: "Online meeting",
			mockResponse: nerResponse{
				Entities: []Entity{},
			},
			expectedError:    false,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedError {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			service := NewNERService(server.URL)
			location, err := service.ExtractLocation(context.Background(), tt.text)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, location)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLocation, location)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	now := time.Now()
	tzUtil := NewTimezoneUtil("UTC")

	tests := []struct {
		name         string
		text         string
		expectedTime time.Time
		expectError  bool
	}{
		{
			name:         "ISO format",
			text:         "2023-12-25T15:30:00Z",
			expectedTime: time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
			expectError:  false,
		},
		{
			name:         "DD/MM/YYYY HH:MM",
			text:         "25/12/2023 15:30",
			expectedTime: time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
			expectError:  false,
		},
		{
			name:         "HH:MM only",
			text:         "15:30",
			expectedTime: time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, time.UTC),
			expectError:  false,
		},
		{
			name:         "with timezone",
			text:         "2023-12-25T15:30:00 EST",
			expectedTime: time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
			expectError:  false,
		},
		{
			name:        "invalid format",
			text:        "not a date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseDateTime(tzUtil, tt.text)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTime.Hour(), parsed.Hour())
				assert.Equal(t, tt.expectedTime.Minute(), parsed.Minute())
				if tt.name == "with timezone" {
					assert.Equal(t, "America/New_York", parsed.Location().String())
				}
			}
		})
	}
}
