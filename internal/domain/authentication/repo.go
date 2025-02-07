package authentication

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gmhafiz/scs/v2"
)

type repo struct {
	db      *sql.DB
	session *scs.SessionManager
}

func NewRepo(db *sql.DB, session *scs.SessionManager) Repo {
	return &repo{
		db:      db,
		session: session,
	}
}

func (r *repo) Register(ctx context.Context, firstName, lastName, email, password string) error {
	query := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, firstName, lastName, email, password)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) Login(ctx context.Context, req LoginRequest) (*User, bool, error) {
	var user User
	query := `
		SELECT id, first_name, last_name, email, password
		FROM users
		WHERE email = $1
	`
	err := r.db.QueryRowContext(ctx, query, req.Email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &user, true, nil
}

func (r *repo) Logout(ctx context.Context, userID uint64) (bool, error) {
	query := `
		DELETE FROM sessions
		WHERE user_id = $1
	`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (r *repo) Csrf(ctx context.Context) (string, error) {
	// TODO: Implement CSRF token generation and storage
	return "", nil
}
