package postgresstore

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	// Thiết lập biến môi trường cho test
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASS", "password")
	os.Setenv("DB_NAME", "go8_test_db")
	os.Setenv("DB_SSL_MODE", "disable")

	// Tạo connection string đến default database
	defaultDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Kết nối đến default database để tạo test database
	defaultDB, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		log.Fatal(err)
	}

	// Tạo test database
	_, err = defaultDB.Exec("DROP DATABASE IF EXISTS go8_test_db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = defaultDB.Exec("CREATE DATABASE go8_test_db")
	if err != nil {
		log.Fatal(err)
	}
	defaultDB.Close()

	// Tạo connection string đến test database
	testDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Kết nối đến test database
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Tạo các bảng cần thiết
	ctx := context.Background()
	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS users
(
    id          bigint generated always as identity primary key,
    first_name  text,
    middle_name text,
    last_name   text,
    email       text unique,
    password    text,
    verified_at timestamptz
);`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS sessions
(
    token   TEXT PRIMARY KEY,
    user_id BIGINT      NOT NULL CONSTRAINT session_user_fk REFERENCES users ON DELETE CASCADE,
    data    BYTEA       NOT NULL,
    expiry  TIMESTAMPTZ NOT NULL
);`)
	if err != nil {
		log.Fatal(err)
	}

	// Thêm dữ liệu test
	_, err = db.ExecContext(ctx, `
INSERT INTO users (email, password, verified_at) 
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING`,
		"admin@example.com", "test", time.Now(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Chạy test
	code := m.Run()

	// Cleanup
	db.Close()
	defaultDB, err = sql.Open("postgres", defaultDSN)
	if err != nil {
		log.Fatal(err)
	}
	_, err = defaultDB.Exec("DROP DATABASE IF EXISTS go8_test_db")
	if err != nil {
		log.Fatal(err)
	}
	defaultDB.Close()

	os.Unsetenv("DB_DRIVER")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASS")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SSL_MODE")

	os.Exit(code)
}
