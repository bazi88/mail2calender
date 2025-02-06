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
				if err := json.NewEncoder(w).Encode(tt.mockResponse); err != nil {
					t.Errorf("failed to encode response: %v", err)
				}
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
						Text:       "tomorrow",
						Label:      "DATE",
						Start:      8,
						End:        16,
						Confidence: 0.95,
					},
					{
						Text:       "2pm",
						Label:      "TIME",
						Start:      20,
						End:        23,
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
			name: "specific date format",
			text: "Meeting on 2024-02-15 at 15:30",
			mockResponse: nerResponse{
				Entities: []Entity{
					{
						Text:       "2024-02-15",
						Label:      "DATE",
						Start:      11,
						End:        21,
						Confidence: 0.95,
					},
					{
						Text:       "15:30",
						Label:      "TIME",
						Start:      25,
						End:        30,
						Confidence: 0.95,
					},
				},
			},
			expectedError: false,
			expectedDates: []time.Time{
				time.Date(2024, 2, 15, 15, 30, 0, 0, time.Local),
			},
		},
		{
			name: "no time entities",
			text: "Just a regular text",
			mockResponse: nerResponse{
				Entities: []Entity{
					{
						Text:       "regular",
						Label:      "ADJ",
						Start:      7,
						End:        14,
						Confidence: 0.9,
					},
				},
			},
			expectedError: true,
			expectedDates: nil,
		},
		{
			name: "server error",
			text: "Meeting tomorrow",
			mockResponse: nerResponse{
				Entities: nil,
			},
			expectedError: true,
			expectedDates: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.name == "server error" {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Verify request
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/v1/extract", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Return mock response
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(tt.mockResponse); err != nil {
					t.Errorf("failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			service := NewNERService(server.URL)
			dates, err := service.ExtractDateTime(context.Background(), tt.text)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, dates)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, dates)
			assert.Len(t, dates, len(tt.expectedDates))

			if len(dates) > 0 && len(tt.expectedDates) > 0 {
				minLen := len(dates)
				if len(tt.expectedDates) < minLen {
					minLen = len(tt.expectedDates)
				}
				for i := 0; i < minLen; i++ {
					assert.Equal(t, tt.expectedDates[i].Year(), dates[i].Year())
					assert.Equal(t, tt.expectedDates[i].Month(), dates[i].Month())
					assert.Equal(t, tt.expectedDates[i].Day(), dates[i].Day())
					assert.Equal(t, tt.expectedDates[i].Hour(), dates[i].Hour())
					assert.Equal(t, tt.expectedDates[i].Minute(), dates[i].Minute())
				}
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	tzUtil := NewTimezoneUtil("Asia/Ho_Chi_Minh")
	now := time.Now()

	tests := []struct {
		name         string
		text         string
		expectedTime time.Time
		expectError  bool
	}{
		{
			name:         "ISO format",
			text:         "2024-02-15T15:30:00+07:00",
			expectedTime: time.Date(2024, 2, 15, 15, 30, 0, 0, time.Local),
			expectError:  false,
		},
		{
			name:         "DD/MM/YYYY HH:MM",
			text:         "15/02/2024 15:30",
			expectedTime: time.Date(2024, 2, 15, 15, 30, 0, 0, time.Local),
			expectError:  false,
		},
		{
			name:         "HH:MM only",
			text:         "15:30",
			expectedTime: time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, time.Local),
			expectError:  false,
		},
		{
			name:         "natural language",
			text:         "3pm",
			expectedTime: time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.Local),
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
				return
			}

			assert.NoError(t, err)
			if !tt.expectedTime.IsZero() {
				assert.Equal(t, tt.expectedTime.Year(), parsed.Year())
				assert.Equal(t, tt.expectedTime.Month(), parsed.Month())
				assert.Equal(t, tt.expectedTime.Day(), parsed.Day())
				assert.Equal(t, tt.expectedTime.Hour(), parsed.Hour())
				assert.Equal(t, tt.expectedTime.Minute(), parsed.Minute())
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
		{
			name:             "server error",
			text:             "Meeting at office",
			mockResponse:     nerResponse{},
			expectedError:    true,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.name == "server error" {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if err := json.NewEncoder(w).Encode(tt.mockResponse); err != nil {
					t.Errorf("failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			service := NewNERService(server.URL)
			location, err := service.ExtractLocation(context.Background(), tt.text)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, location)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLocation, location)
		})
	}
}
