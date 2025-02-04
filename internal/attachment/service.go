package attachment

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

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
		return err
	}
	defer f.Close()

	// Read file header for MIME type detection
	head := make([]byte, 261)
	if _, err := f.Read(head); err != nil && err != io.EOF {
		return err
	}

	kind, err := filetype.Match(head)
	if err != nil {
		return err
	}

	allowed := false
	for _, t := range AllowedTypes {
		if strings.HasPrefix(t, kind.MIME.Value) {
			allowed = true
			break
		}
	}

	if !allowed {
		return errors.New("file type not allowed")
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
