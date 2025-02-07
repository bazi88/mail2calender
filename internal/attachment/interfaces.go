package attachment

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

// Storage defines the interface for file storage operations
type Storage interface {
	Save(ctx context.Context, data []byte, ext string) (string, error)
	Get(ctx context.Context, fileID string) ([]byte, string, error)
	Delete(ctx context.Context, fileID string) error
}

// MinioClientInterface defines the interface for minio client operations
type MinioClientInterface interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
}

// VirusScanner định nghĩa interface cho việc quét virus
type VirusScanner interface {
	Scan(data []byte) (bool, error)
}
