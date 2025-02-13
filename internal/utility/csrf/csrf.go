package csrf

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/cespare/xxhash/v2"
)

// Package csrf cung cấp các chức năng xử lý CSRF token

// ValidToken kiểm tra xem CSRF token có hợp lệ hay không
func ValidToken(ctx context.Context, db *sql.DB, token string) bool {
	tokenHash, err := sum(token)
	if err != nil {
		return false
	}

	var exists bool
	row := db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT token FROM sessions 
				WHERE token = $1 
				  AND current_timestamp < expiry
			) `, tokenHash)
	if err = row.Scan(&exists); err != nil {
		return false
	}
	return exists
}

// ValidAndDeleteToken xóa token khỏi store nếu token hợp lệ.
// Hữu ích cho việc sử dụng token một lần.
func ValidAndDeleteToken(ctx context.Context, db *sql.DB, token string) error {
	tokenHash, err := sum(token)
	if err != nil {
		return nil
	}

	res, err := db.ExecContext(ctx, `
		DELETE FROM sessions WHERE token = $1 AND current_timestamp < expiry
	`, tokenHash)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.New("token not found")
	}

	if rowsAffected != 1 {
		return errors.New("no csrf token was found")
	}
	return nil
}

// sum tính toán hash của token
func sum(token string) (string, error) {
	h := xxhash.New()
	if _, err := h.Write([]byte(token)); err != nil {
		return "", err
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum), nil
}
