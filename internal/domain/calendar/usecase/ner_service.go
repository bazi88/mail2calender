package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	var dateEntity, timeEntity *Entity

	// First pass: find DATE and TIME entities
	for _, entity := range entities {
		if strings.EqualFold(entity.Label, "DATE") {
			dateEntity = &entity
		} else if strings.EqualFold(entity.Label, "TIME") {
			timeEntity = &entity
		}
	}

	// If we have both date and time, combine them
	if dateEntity != nil && timeEntity != nil {
		// Parse date first
		dateTime, err := parseDateTime(s.tzUtil, dateEntity.Text)
		if err != nil {
			return nil, err
		}

		// Parse time and combine with date
		timeOnly, err := parseDateTime(s.tzUtil, timeEntity.Text)
		if err != nil {
			return nil, err
		}

		// Combine date and time
		combinedTime := time.Date(
			dateTime.Year(),
			dateTime.Month(),
			dateTime.Day(),
			timeOnly.Hour(),
			timeOnly.Minute(),
			0, 0,
			dateTime.Location(),
		)
		dates = append(dates, combinedTime)
	} else {
		// If we only have one entity, try to parse it
		for _, entity := range entities {
			if strings.EqualFold(entity.Label, "TIME") || strings.EqualFold(entity.Label, "DATE") {
				t, err := parseDateTime(s.tzUtil, entity.Text)
				if err == nil {
					dates = append(dates, t)
				}
			}
		}
	}

	if len(dates) == 0 {
		return nil, fmt.Errorf("no valid dates found in text")
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

	// Xử lý các từ khóa thời gian tự nhiên
	switch strings.ToLower(text) {
	case "tomorrow":
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local), nil
	case "today":
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), nil
	}

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

	// Handle natural language time format (e.g., "3pm")
	if strings.HasSuffix(strings.ToLower(text), "pm") || strings.HasSuffix(strings.ToLower(text), "am") {
		hour := 0
		meridiem := strings.ToLower(text[len(text)-2:])
		hourStr := strings.TrimSuffix(strings.TrimSuffix(text, "pm"), "am")

		if h, err := strconv.Atoi(hourStr); err == nil {
			if meridiem == "pm" && h < 12 {
				hour = h + 12
			} else if meridiem == "am" && h == 12 {
				hour = 0
			} else {
				hour = h
			}
			now := time.Now()
			return time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local), nil
		}
	}

	// Handle HH:MM format
	if strings.Contains(text, ":") {
		parts := strings.Split(text, ":")
		if len(parts) == 2 {
			hour, errHour := strconv.Atoi(parts[0])
			minute, errMin := strconv.Atoi(parts[1])
			if errHour == nil && errMin == nil {
				now := time.Now()
				return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local), nil
			}
		}
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
