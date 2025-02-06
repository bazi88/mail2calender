package attachment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

var allowedExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".txt":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

type MinioStorage struct {
	client     MinioClientInterface
	bucketName string
}

// NewMinioStorage creates a new MinIO storage instance
func NewMinioStorage(client *minio.Client, bucketName string) Storage {
	return &MinioStorage{
		client:     client,
		bucketName: bucketName,
	}
}

// validateFile checks if the file meets size and extension requirements
func (s *MinioStorage) validateFile(data []byte, ext string) error {
	if len(data) > maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxFileSize)
	}
	if !allowedExtensions[strings.ToLower(ext)] {
		return fmt.Errorf("file extension %s is not allowed", ext)
	}
	return nil
}

func (s *MinioStorage) Save(ctx context.Context, data []byte, ext string) (string, error) {
	if err := s.validateFile(data, ext); err != nil {
		return "", err
	}

	objectName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, bytes.NewReader(data), int64(len(data)), opts)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return objectName, nil
}

func (s *MinioStorage) Get(ctx context.Context, fileID string) ([]byte, string, error) {
	obj, err := s.client.GetObject(ctx, s.bucketName, fileID, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	ext := filepath.Ext(fileID)
	return data, ext, nil
}

func (s *MinioStorage) Delete(ctx context.Context, fileID string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, fileID, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// getContentType returns the MIME type based on file extension
func getContentType(ext string) string {
	ext = strings.ToLower(ext)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// Default to application/octet-stream if MIME type is unknown
		return "application/octet-stream"
	}
	return mimeType
}
