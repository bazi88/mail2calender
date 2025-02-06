package attachment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ClamAVScanner struct {
	endpoint string
	timeout  time.Duration
}

func NewClamAVScanner(endpoint string, timeout time.Duration) VirusScanner {
	return &ClamAVScanner{
		endpoint: endpoint,
		timeout:  timeout,
	}
}

func (s *ClamAVScanner) Scan(data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint+"/scan", bytes.NewReader(data))
	if err != nil {
		return false, fmt.Errorf("failed to create scan request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send scan request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// ClamAV returns "OK" if no virus is found
	return string(body) == "OK", nil
}
