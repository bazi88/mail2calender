package attachment

import "context"

// Storage định nghĩa interface cho việc lưu trữ file
type Storage interface {
	Save(ctx context.Context, data []byte, ext string) (string, error)
	Get(ctx context.Context, fileID string) ([]byte, string, error)
	Delete(ctx context.Context, fileID string) error
}

// VirusScanner định nghĩa interface cho việc quét virus
type VirusScanner interface {
	Scan(data []byte) (bool, error)
}
