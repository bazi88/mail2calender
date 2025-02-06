package attachment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

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

	ext = strings.ToLower(ext)
	if !allowedExtensions[ext] {
		return fmt.Errorf("file extension %s is not allowed", ext)
	}
	return nil
}

// Save stores a file in MinIO storage
func (s *MinioStorage) Save(ctx context.Context, data []byte, ext string) (string, error) {
	if err := s.validateFile(data, ext); err != nil {
		return "", err
	}

	fileID := uuid.New().String()
	key := fmt.Sprintf("%s/%s%s", time.Now().Format("2006/01/02"), fileID, ext)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := s.client.PutObject(ctx, s.bucketName, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: getContentType(ext),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	return key, nil
}

// Get retrieves a file from MinIO storage
func (s *MinioStorage) Get(ctx context.Context, fileID string) ([]byte, string, error) {
	ext := strings.ToLower(filepath.Ext(fileID))
	if !allowedExtensions[ext] {
		return nil, "", fmt.Errorf("file extension %s is not allowed", ext)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	object, err := s.client.GetObject(ctx, s.bucketName, fileID, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file from MinIO: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file content: %w", err)
	}

	if len(data) > maxFileSize {
		return nil, "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxFileSize)
	}

	return data, ext, nil
}

// Delete removes a file from MinIO storage
func (s *MinioStorage) Delete(ctx context.Context, fileID string) error {
	ext := strings.ToLower(filepath.Ext(fileID))
	if !allowedExtensions[ext] {
		return fmt.Errorf("file extension %s is not allowed", ext)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := s.client.RemoveObject(ctx, s.bucketName, fileID, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
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
