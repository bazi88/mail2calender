package attachment

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
)

type QuarantineStorage interface {
	Storage
	MarkAsQuarantined(ctx context.Context, fileID string) error
}

type S3Storage struct {
	client           MinioClientInterface
	bucket           string
	quarantineBucket string
}

func (s *S3Storage) SaveWithRetry(ctx context.Context, data []byte, ext string, retries int) (string, error) {
	var lastErr error
	for i := 0; i < retries; i++ {
		id, err := s.Save(ctx, data, ext)
		if err == nil {
			return id, nil
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return "", fmt.Errorf("after %d retries: %w", retries, lastErr)
}

func (s *S3Storage) MarkAsQuarantined(ctx context.Context, fileID string) error {
	// Get the file from main storage
	data, ext, err := s.Get(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file for quarantine: %w", err)
	}

	// Move to quarantine bucket
	_, err = s.client.PutObject(ctx, s.quarantineBucket, fileID+ext, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to move file to quarantine: %w", err)
	}

	// Delete from main storage
	err = s.Delete(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete quarantined file from main storage: %w", err)
	}

	return nil
}
