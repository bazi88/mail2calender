package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Entity represents a named entity from NER service
type Entity struct {
	Text       string  `json:"text"`
	Label      string  `json:"label"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Confidence float64 `json:"confidence"`
}

type nerRequest struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

type nerResponse struct {
	Entities       []Entity `json:"entities"`
	ProcessingTime float64  `json:"processing_time"`
}

// NERService handles communication with the NER microservice
type NERService interface {
	ExtractEntities(ctx context.Context, text string, language string) ([]Entity, error)
	ExtractDateTime(ctx context.Context, text string) ([]time.Time, error)
	ExtractLocation(ctx context.Context, text string) (string, error)
}

type nerServiceImpl struct {
	client  *http.Client
	baseURL string
	tzUtil  *TimezoneUtil
}

func NewNERService(baseURL string) NERService {
	return &nerServiceImpl{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: baseURL,
		tzUtil:  NewTimezoneUtil("Asia/Ho_Chi_Minh"), // Default to Vietnam timezone
	}
}

func (s *nerServiceImpl) ExtractEntities(ctx context.Context, text string, language string) ([]Entity, error) {
	reqBody := nerRequest{
		Text:     text,
		Language: language,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/v1/extract", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NER service returned status: %d", resp.StatusCode)
	}

	var result nerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result.Entities, nil
}

func (s *nerServiceImpl) ExtractDateTime(ctx context.Context, text string) ([]time.Time, error) {
	entities, err := s.ExtractEntities(ctx, text, "vi") // Default to Vietnamese
	if err != nil {
		return nil, err
	}

	var dates []time.Time
	for _, entity := range entities {
		if entity.Label == "DATE" || entity.Label == "TIME" {
			// Parse date/time text using various formats
			t, err := parseDateTime(s.tzUtil, entity.Text)
			if err == nil {
				dates = append(dates, t)
			}
		}
	}

	return dates, nil
}

func (s *nerServiceImpl) ExtractLocation(ctx context.Context, text string) (string, error) {
	entities, err := s.ExtractEntities(ctx, text, "vi")
	if err != nil {
		return "", err
	}

	// Look for location entities with highest confidence
	var bestLocation string
	var bestConfidence float64

	for _, entity := range entities {
		if entity.Label == "LOC" && entity.Confidence > bestConfidence {
			bestLocation = entity.Text
			bestConfidence = entity.Confidence
		}
	}

	return bestLocation, nil
}

// parseDateTime attempts to parse date/time text in various formats
func parseDateTime(tzUtil *TimezoneUtil, text string) (time.Time, error) {
	text = strings.TrimSpace(text)

	// Check for timezone abbreviation in the text
	var timezoneName string
	for tzAbbr := range getTimezoneAbbreviations() {
		if strings.Contains(strings.ToUpper(text), tzAbbr) {
			timezoneName = tzUtil.GuessTimezone(tzAbbr)
			text = strings.ReplaceAll(strings.ToUpper(text), tzAbbr, "")
			text = strings.TrimSpace(text)
			break
		}
	}

	if timezoneName == "" {
		timezoneName = tzUtil.defaultTimezone
	}

	// Common formats to try
	formats := []string{
		"2006-01-02T15:04:05Z07:00",       // ISO
		"2006-01-02T15:04:05-0700",        // ISO with numeric zone
		"Monday, January 2, 2006 3:04 PM", // Long format
		"Mon, 02 Jan 2006 15:04:05 -0700", // RFC822Z
		"02/01/2006 15:04:05",             // Full time
		"02/01/2006 15:04",                // DD/MM/YYYY HH:MM
		"02-01-2006 15:04",                // DD-MM-YYYY HH:MM
		"2006/01/02 15:04",                // YYYY/MM/DD HH:MM
		"02/01/2006",                      // DD/MM/YYYY
		"02-01-2006",                      // DD-MM-YYYY
		"15:04 02/01/2006",                // HH:MM DD/MM/YYYY
		"3:04PM",                          // Local time 12h
		"15:04",                           // 24h time
		"January 2, 2006",                 // Month D, YYYY
		"Jan 2, 2006",                     // Mon D, YYYY
		"2006-01-02",                      // YYYY-MM-DD
	}

	// Try each format
	for _, format := range formats {
		if t, err := tzUtil.ParseTimeInTimezone(text, format, timezoneName); err == nil {
			// If no year specified, use current year
			if t.Year() == 0 {
				now := time.Now()
				t = time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse datetime: %s", text)
}

func getTimezoneAbbreviations() map[string]struct{} {
	return map[string]struct{}{
		"EST": {}, "EDT": {},
		"CST": {}, "CDT": {},
		"MST": {}, "MDT": {},
		"PST": {}, "PDT": {},
		"GMT": {}, "UTC": {},
		"ICT": {}, "JST": {},
		"IST": {}, "AEST": {},
	}
}
