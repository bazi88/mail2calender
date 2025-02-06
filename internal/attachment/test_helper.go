package attachment

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func setupTestMinio(t *testing.T) (*minio.Client, string, func()) {
	if err := godotenv.Load("test.env"); err != nil {
		t.Fatalf("Error loading test.env file: %v", err)
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := os.Getenv("MINIO_BUCKET")

	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		t.Fatalf("Error creating MinIO client: %v", err)
	}

	// Create test bucket if it doesn't exist
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		t.Fatalf("Error checking bucket existence: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			t.Fatalf("Error creating test bucket: %v", err)
		}
	}

	cleanup := func() {
		// Remove all objects in the bucket
		objectsCh := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
			Recursive: true,
		})
		for obj := range objectsCh {
			if obj.Err != nil {
				log.Printf("Error listing objects: %v", obj.Err)
				continue
			}
			err := minioClient.RemoveObject(context.Background(), bucketName, obj.Key, minio.RemoveObjectOptions{})
			if err != nil {
				log.Printf("Error removing object %s: %v", obj.Key, err)
			}
		}
	}

	return minioClient, bucketName, cleanup
}
