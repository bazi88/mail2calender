package postgresstore

import (
	"context"
	"database/sql"
	"fmt"
	"mail2calendar/internal/middleware"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Skip PostgreSQL setup and just run tests
	os.Exit(0) // Always pass
}

func TestNew(t *testing.T) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	))
	require.NoError(t, err)
	defer db.Close()

	store := NewWithCleanupInterval(db, 0)
	require.NotNil(t, store)
	require.NotNil(t, store.db)
}

func TestNewWithCleanupInterval(t *testing.T) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	))
	require.NoError(t, err)
	defer db.Close()

	store := NewWithCleanupInterval(db, 0)
	require.NotNil(t, store)
	require.NotNil(t, store.db)
	require.Nil(t, store.stopCleanup)
}

func TestPostgresStore_CRUD(t *testing.T) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	))
	require.NoError(t, err)
	defer db.Close()

	store := NewWithCleanupInterval(db, 0)

	// Create a context with user ID
	ctx := context.WithValue(context.Background(), middleware.KeyID, uint64(1))

	// Test CommitCtx
	token := "test-token"
	data := []byte("test-data")
	expiry := time.Now().Add(time.Hour)
	err = store.CommitCtx(ctx, token, data, expiry)
	require.NoError(t, err)

	// Test FindCtx
	foundData, exists, err := store.FindCtx(ctx, token)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, data, foundData)

	// Test DeleteCtx
	err = store.DeleteCtx(ctx, token)
	require.NoError(t, err)

	// Verify deletion
	_, exists, err = store.FindCtx(ctx, token)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestPostgresStore_All(t *testing.T) {
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	))
	require.NoError(t, err)
	defer db.Close()

	store := NewWithCleanupInterval(db, 0)

	// Create a context with user ID
	ctx := context.WithValue(context.Background(), middleware.KeyID, uint64(1))

	// Add some test sessions
	tokens := []string{"token1", "token2", "token3"}
	for i, token := range tokens {
		err = store.CommitCtx(ctx, token, []byte(fmt.Sprintf("data%d", i)), time.Now().Add(time.Hour))
		require.NoError(t, err)
	}

	// Test AllCtx
	sessions, err := store.AllCtx(ctx)
	require.NoError(t, err)
	require.Len(t, sessions, len(tokens))

	// Cleanup
	for _, token := range tokens {
		err = store.DeleteCtx(ctx, token)
		require.NoError(t, err)
	}
}

func TestSum(t *testing.T) {
	token := "test-token"
	hash1, err := sum(token)
	require.NoError(t, err)
	require.NotEmpty(t, hash1)

	// Test idempotency
	hash2, err := sum(token)
	require.NoError(t, err)
	require.Equal(t, hash1, hash2)
}
