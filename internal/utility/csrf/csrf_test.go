package csrf

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestValidToken(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		mockSetup func(mock sqlmock.Sqlmock)
		expected  bool
	}{
		{
			name:  "valid token",
			token: "valid-token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT EXISTS\(.*\)`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expected: true,
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT EXISTS\(.*\)`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expected: false,
		},
		{
			name:  "database error",
			token: "error-token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT EXISTS\(.*\)`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock db
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			// Setup expectations
			tt.mockSetup(mock)

			// Call function
			result := ValidToken(context.Background(), db, tt.token)

			// Assert result
			assert.Equal(t, tt.expected, result)

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestValidAndDeleteToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name:  "valid token",
			token: "valid-token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM sessions`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
		{
			name:  "token not found",
			token: "invalid-token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM sessions`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedError: errors.New("no csrf token was found"),
		},
		{
			name:  "database error",
			token: "error-token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM sessions`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock db
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			// Setup expectations
			tt.mockSetup(mock)

			// Call function
			err = ValidAndDeleteToken(context.Background(), db, tt.token)

			// Assert error
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "valid token",
			token:       "test-token",
			expectError: false,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := sum(tt.token)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
			}

			// Verify hash is consistent
			hash2, err := sum(tt.token)
			assert.NoError(t, err)
			assert.Equal(t, hash, hash2)
		})
	}
}
