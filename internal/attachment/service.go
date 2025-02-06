package attachment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/h2non/filetype"
)

var (
	MaxFileSize  = int64(10 * 1024 * 1024) // 10MB
	AllowedTypes = []string{
		"image/jpeg", "image/png", "image/gif",
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
)

type Service interface {
	Upload(ctx context.Context, file *multipart.FileHeader) (string, error)
	Download(ctx context.Context, fileID string) ([]byte, string, error)
	Delete(ctx context.Context, fileID string) error
	ValidateFile(file *multipart.FileHeader) error
}

type service struct {
	storage Storage
	scanner VirusScanner
}

func NewService(storage Storage, scanner VirusScanner) Service {
	return &service{
		storage: storage,
		scanner: scanner,
	}
}

func (s *service) ValidateFile(file *multipart.FileHeader) error {
	if file.Size > MaxFileSize {
		return errors.New("file size exceeds maximum limit")
	}

	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Read file content for validation
	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Check file type
	kind, err := filetype.Match(data)
	if err != nil || kind == filetype.Unknown {
		return errors.New("invalid or unsupported file type")
	}

	// Validate MIME type
	validType := false
	for _, allowedType := range AllowedTypes {
		if kind.MIME.Value == allowedType {
			validType = true
			break
		}
	}

	if !validType {
		return errors.New("file type not allowed")
	}

	// Scan for viruses
	clean, err := s.scanner.Scan(data)
	if err != nil {
		return fmt.Errorf("virus scan failed: %w", err)
	}

	if !clean {
		return errors.New("file appears to be infected")
	}

	return nil
}

func (s *service) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	if err := s.ValidateFile(file); err != nil {
		return "", err
	}

	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		return "", err
	}

	// Scan file for viruses
	clean, err := s.scanner.Scan(buf.Bytes())
	if err != nil {
		return "", err
	}
	if !clean {
		return "", errors.New("file contains malware")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	return s.storage.Save(ctx, buf.Bytes(), ext)
}

func (s *service) Download(ctx context.Context, fileID string) ([]byte, string, error) {
	return s.storage.Get(ctx, fileID)
}

func (s *service) Delete(ctx context.Context, fileID string) error {
	return s.storage.Delete(ctx, fileID)
}
