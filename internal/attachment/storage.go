package attachment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// FileInfo represents metadata about a stored file
type FileInfo struct {
	ID        string
	CreatedAt time.Time
}

type QuarantineStorage interface {
	Storage
	MarkAsQuarantined(ctx context.Context, fileID string) error
}

type S3Storage struct {
	client           MinioClientInterface
	bucket           string
	quarantineBucket string
}

func NewS3Storage(client MinioClientInterface, bucket, quarantineBucket string) *S3Storage {
	return &S3Storage{
		client:           client,
		bucket:           bucket,
		quarantineBucket: quarantineBucket,
	}
}

func (s *S3Storage) Save(ctx context.Context, data []byte, ext string) (string, error) {
	fileID := uuid.New().String()
	objectName := fileID
	if ext != "" {
		objectName = fileID + "." + ext
	}
	_, err := s.client.PutObject(ctx, s.bucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	return fileID, nil
}

func (s *S3Storage) Get(ctx context.Context, id string) ([]byte, string, error) {
	// List objects with prefix to find the file with extension
	pattern := id + ".*"
	var objectName string
	var ext string

	objects := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix: id,
	})
	for object := range objects {
		if object.Err != nil {
			return nil, "", fmt.Errorf("failed to list files: %w", object.Err)
		}
		if matched, _ := filepath.Match(pattern, object.Key); matched {
			objectName = object.Key
			ext = filepath.Ext(object.Key)
			break
		}
	}

	if objectName == "" {
		return nil, "", fmt.Errorf("file not found: %s", id)
	}

	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file: %w", err)
	}
	defer func() {
		if cerr := obj.Close(); cerr != nil {
			err = fmt.Errorf("failed to close object: %v", cerr)
		}
	}()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}
	return data, ext, nil
}

func (s *S3Storage) Delete(ctx context.Context, id string) error {
	// List objects with prefix to find the file with extension
	pattern := id + ".*"
	var objectName string

	objects := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix: id,
	})
	for object := range objects {
		if object.Err != nil {
			return fmt.Errorf("failed to list files: %w", object.Err)
		}
		if matched, _ := filepath.Match(pattern, object.Key); matched {
			objectName = object.Key
			break
		}
	}

	if objectName == "" {
		return fmt.Errorf("file not found: %s", id)
	}

	err := s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *S3Storage) ListFiles(ctx context.Context) ([]FileInfo, error) {
	var files []FileInfo

	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{})
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list files: %w", object.Err)
		}
		files = append(files, FileInfo{
			ID:        object.Key,
			CreatedAt: object.LastModified,
		})
	}

	return files, nil
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
	objectName := fileID
	if ext != "" {
		objectName = fileID + ext
	}
	_, err = s.client.PutObject(ctx, s.quarantineBucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
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
