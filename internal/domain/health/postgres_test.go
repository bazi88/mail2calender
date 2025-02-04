package health

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNewRepo(t *testing.T) {
	repo := NewRepo((*sqlx.DB)(nil))
	assert.NotNil(t, repo)
}

func TestRepository_Readiness(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		expectedError error
	}{
		{
			name:          "successful ping",
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "database error",
			mockError:     errors.New("connection error"),
			expectedError: errors.New("connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock database with ping monitoring enabled
			mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			assert.NoError(t, err)
			defer mockDB.Close()

			// Convert to sqlx.DB
			db := sqlx.NewDb(mockDB, "sqlmock")
			defer db.Close()

			// Set up expectations
			if tt.mockError != nil {
				mock.ExpectPing().WillReturnError(tt.mockError)
			} else {
				mock.ExpectPing()
			}

			repo := NewRepo(db)
			err = repo.Readiness()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
