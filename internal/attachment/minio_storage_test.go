package attachment

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMinioClient struct {
	mock.Mock
}

func (m *mockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (info minio.UploadInfo, err error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	return args.Get(0).(minio.UploadInfo), args.Error(1)
}

func (m *mockMinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (object *minio.Object, err error) {
	args := m.Called(ctx, bucketName, objectName, opts)
	if obj := args.Get(0); obj != nil {
		return obj.(*minio.Object), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMinioClient) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Error(0)
}

func (m *mockMinioClient) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	args := m.Called(ctx, bucketName, opts)
	return args.Get(0).(<-chan minio.ObjectInfo)
}

func TestMinioStorage_Save(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		ext         string
		setupMock   func(*mockMinioClient)
		wantErr     bool
		expectedErr string
	}{
		{
			name: "successful save",
			data: []byte("test data"),
			ext:  ".pdf",
			setupMock: func(m *mockMinioClient) {
				m.On("PutObject",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(minio.UploadInfo{}, nil)
			},
			wantErr: false,
		},
		{
			name: "file too large",
			data: make([]byte, maxFileSize+1),
			ext:  ".pdf",
			setupMock: func(m *mockMinioClient) {
				// No mock needed as it should fail before calling PutObject
			},
			wantErr:     true,
			expectedErr: "file size exceeds maximum allowed size",
		},
		{
			name: "invalid extension",
			data: []byte("test data"),
			ext:  ".invalid",
			setupMock: func(m *mockMinioClient) {
				// No mock needed as it should fail before calling PutObject
			},
			wantErr:     true,
			expectedErr: "file extension .invalid is not allowed",
		},
		{
			name: "context timeout",
			data: []byte("test data"),
			ext:  ".pdf",
			setupMock: func(m *mockMinioClient) {
				m.On("PutObject",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).After(2*time.Second).Return(minio.UploadInfo{}, context.DeadlineExceeded)
			},
			wantErr:     true,
			expectedErr: "failed to upload file to MinIO: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockMinioClient)
			tt.setupMock(mockClient)

			storage := &MinioStorage{
				client:     mockClient,
				bucketName: "test-bucket",
			}

			ctx := context.Background()
			if tt.name == "context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Second)
				defer cancel()
			}

			key, err := storage.Save(ctx, tt.data, tt.ext)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Empty(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, key)
				assert.Contains(t, key, time.Now().Format("2006/01/02"))
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestMinioStorage_Get(t *testing.T) {
	tests := []struct {
		name        string
		fileID      string
		setupMock   func(*mockMinioClient)
		wantErr     bool
		expectedErr string
	}{
		{
			name:   "invalid extension",
			fileID: "test.invalid",
			setupMock: func(m *mockMinioClient) {
				// No mock needed as it should fail before calling GetObject
			},
			wantErr:     true,
			expectedErr: "file extension .invalid is not allowed",
		},
		{
			name:   "minio error",
			fileID: "test.pdf",
			setupMock: func(m *mockMinioClient) {
				m.On("GetObject",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, assert.AnError)
			},
			wantErr:     true,
			expectedErr: "failed to get file from MinIO",
		},
		{
			name:   "context timeout",
			fileID: "test.pdf",
			setupMock: func(m *mockMinioClient) {
				m.On("GetObject",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).After(2*time.Second).Return(nil, context.DeadlineExceeded)
			},
			wantErr:     true,
			expectedErr: "failed to get file from MinIO: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockMinioClient)
			tt.setupMock(mockClient)

			storage := &MinioStorage{
				client:     mockClient,
				bucketName: "test-bucket",
			}

			ctx := context.Background()
			if tt.name == "context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Second)
				defer cancel()
			}

			data, ext, err := storage.Get(ctx, tt.fileID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, data)
				assert.Empty(t, ext)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				assert.NotEmpty(t, ext)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestMinioStorage_Delete(t *testing.T) {
	tests := []struct {
		name        string
		fileID      string
		setupMock   func(*mockMinioClient)
		wantErr     bool
		expectedErr string
	}{
		{
			name:   "successful delete",
			fileID: "test.pdf",
			setupMock: func(m *mockMinioClient) {
				m.On("RemoveObject",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "invalid extension",
			fileID: "test.invalid",
			setupMock: func(m *mockMinioClient) {
				// No mock needed as it should fail before calling RemoveObject
			},
			wantErr:     true,
			expectedErr: "file extension .invalid is not allowed",
		},
		{
			name:   "minio error",
			fileID: "test.pdf",
			setupMock: func(m *mockMinioClient) {
				m.On("RemoveObject",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(assert.AnError)
			},
			wantErr:     true,
			expectedErr: "failed to delete file from MinIO",
		},
		{
			name:   "context timeout",
			fileID: "test.pdf",
			setupMock: func(m *mockMinioClient) {
				m.On("RemoveObject",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).After(2 * time.Second).Return(context.DeadlineExceeded)
			},
			wantErr:     true,
			expectedErr: "failed to delete file from MinIO: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockMinioClient)
			tt.setupMock(mockClient)

			storage := &MinioStorage{
				client:     mockClient,
				bucketName: "test-bucket",
			}

			ctx := context.Background()
			if tt.name == "context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Second)
				defer cancel()
			}

			err := storage.Delete(ctx, tt.fileID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected string
	}{
		{
			name:     "pdf extension",
			ext:      ".pdf",
			expected: "application/pdf",
		},
		{
			name:     "jpg extension",
			ext:      ".jpg",
			expected: "image/jpeg",
		},
		{
			name:     "unknown extension",
			ext:      ".unknown",
			expected: "application/octet-stream",
		},
		{
			name:     "empty extension",
			ext:      "",
			expected: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getContentType(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMinioStorage_Integration(t *testing.T) {
	// Skip test if MinIO server is not running
	client, err := minio.New("localhost:9000", &minio.Options{
		Secure: false,
	})
	if err != nil {
		t.Skip("MinIO server is not running")
		return
	}

	// Try to connect to MinIO server
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.ListBuckets(ctx)
	if err != nil {
		t.Skip("Could not connect to MinIO server:", err)
		return
	}

	// Rest of the test...
	bucketName := "test-bucket"

	// Check if bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		t.Fatalf("Error checking bucket existence: %v", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			t.Fatalf("Error creating bucket: %v", err)
		}
	}

	// Create storage instance
	storage := NewMinioStorage(client, bucketName)

	// Test Save
	testData := []byte("test data")
	fileID, err := storage.Save(ctx, testData, ".txt")
	if err != nil {
		t.Fatalf("Error saving file: %v", err)
	}

	// Test Get
	retrievedData, ext, err := storage.Get(ctx, fileID)
	if err != nil {
		t.Fatalf("Error getting file: %v", err)
	}
	if string(retrievedData) != string(testData) {
		t.Errorf("Retrieved data does not match: got %s, want %s", retrievedData, testData)
	}
	if ext != ".txt" {
		t.Errorf("Retrieved extension does not match: got %s, want .txt", ext)
	}

	// Test Delete
	err = storage.Delete(ctx, fileID)
	if err != nil {
		t.Fatalf("Error deleting file: %v", err)
	}

	// Verify file is deleted
	_, _, err = storage.Get(ctx, fileID)
	if err == nil {
		t.Error("Expected error getting deleted file")
	}

	// Clean up bucket
	err = client.RemoveBucket(ctx, bucketName)
	if err != nil {
		t.Fatalf("Error removing bucket: %v", err)
	}
}
