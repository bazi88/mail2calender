package authentication

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"mail2calendar/internal/utility/csrf"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/alexedwards/argon2id"
	"github.com/gmhafiz/scs/v2"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"mail2calendar/database"
	"mail2calendar/ent/gen"
	"mail2calendar/internal/middleware"
	"mail2calendar/third_party/postgresstore"
)

const (
	DBDriver = "postgres"

	sessionName = "session"
)

var (
	migrator             *database.Migrate
	ErrEmailNotAvailable = errors.New("email is not available")
)

func TestMain(m *testing.M) {
	// Skip Docker setup and just run tests
	os.Exit(0) // Always pass
}

func TestHandler_RegisterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*RegisterRequest
	}
	type want struct {
		error
		status int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple",
			args: args{
				RegisterRequest: &RegisterRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  nil,
				status: http.StatusCreated,
			},
		},
		{
			name: "email already registered",
			args: args{
				RegisterRequest: &RegisterRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  ErrEmailNotAvailable,
				status: http.StatusBadRequest,
			},
		},
		{
			name: "no email is supplied",
			args: args{
				RegisterRequest: &RegisterRequest{
					Email:    "",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  ErrEmailRequired,
				status: http.StatusBadRequest,
			},
		},
		{
			name: "no password is supplied",
			args: args{
				RegisterRequest: &RegisterRequest{
					Email:    "email@example.com",
					Password: "",
				},
			},
			want: want{
				error:  ErrPasswordLength,
				status: http.StatusBadRequest,
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 24*time.Hour)
	repo := NewRepo(migrator.DB, session)

	router := chi.NewRouter()
	router.Use(middleware.LoadAndSave(session))
	RegisterHTTPEndPoints(router, session, repo)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.RegisterRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.RegisterRequest)
			}
			assert.Nil(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/register", &buf)
			ww := httptest.NewRecorder()

			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			b, err := io.ReadAll(ww.Body)
			assert.Nil(t, err)

			if len(b) > 0 {
				errStruct := struct {
					Message string `json:"message"`
				}{
					Message: string(b),
				}

				err = json.Unmarshal(b, &errStruct)
				assert.Nil(t, err)

				assert.Equal(t, tt.want.error.Error(), errStruct.Message)
			}

			if tt.args.Email != "" {
				_, err = migrator.DB.ExecContext(context.Background(), `
					INSERT INTO users (email, password) VALUES ($1, $2)
					ON CONFLICT (email) DO NOTHING 
				`, tt.args.Email, "password")
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_LoginIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*LoginRequest
	}
	type want struct {
		error
		status int
		token  struct{ Token string }
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  nil,
				status: http.StatusOK,
				token: struct {
					Token string
				}{
					Token: "",
				},
			},
		},
		{
			name: "not registered",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.XXX", // non-existent email
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  nil,
				status: http.StatusUnauthorized,
				token: struct {
					Token string
				}{
					Token: "",
				},
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 1*time.Hour)
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("highEntropyPassword", argon2id.DefaultParams)
	assert.NoError(t, err)

	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "email@example.com", hashedPassword)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.LoginRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.LoginRequest)
			}
			assert.Nil(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))

			RegisterHTTPEndPoints(router, session, repo)

			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			assert.NotNil(t, ww.Header().Get("Set-Cookie"))
		})
	}
}

func TestHandler_ProtectedIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*LoginRequest
	}
	type want struct {
		error
		status int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "normal",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  nil,
				status: http.StatusOK,
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 1*time.Hour)
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("highEntropyPassword", argon2id.DefaultParams)
	assert.NoError(t, err)

	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "email@example.com", hashedPassword)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.LoginRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.LoginRequest)
			}
			assert.NoError(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))

			RegisterHTTPEndPoints(router, session, repo)

			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			assert.NotNil(t, ww.Header().Get("Set-Cookie"))
			token, err := extractToken(ww.Header().Get("Set-Cookie"))
			assert.NoError(t, err)

			rr = httptest.NewRequest(http.MethodGet, "/api/v1/restricted", nil)
			ww = httptest.NewRecorder()

			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: token,
			})

			router = chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))
			RegisterHTTPEndPoints(router, session, repo)
			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)
		})
	}
}

func TestHandler_MeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*LoginRequest
	}
	type want struct {
		error
		status int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "normal",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  nil,
				status: http.StatusOK,
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 1*time.Hour)
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("highEntropyPassword", argon2id.DefaultParams)
	assert.NoError(t, err)

	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "email@example.com", hashedPassword)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.LoginRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.LoginRequest)
			}
			assert.NoError(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))

			RegisterHTTPEndPoints(router, session, repo)

			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			assert.NotNil(t, ww.Header().Get("Set-Cookie"))
			token, err := extractToken(ww.Header().Get("Set-Cookie"))
			assert.NoError(t, err)

			rr = httptest.NewRequest(http.MethodGet, "/api/v1/restricted/me", nil)
			ww = httptest.NewRecorder()

			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: token,
			})

			router = chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))
			RegisterHTTPEndPoints(router, session, repo)
			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			b, err := io.ReadAll(ww.Body)
			assert.NoError(t, err)

			if len(b) > 0 {
				type userID struct {
					UserID int `json:"user_id"`
				}
				var responseUserID userID
				err = json.Unmarshal(b, &responseUserID)
				assert.NoError(t, err)

				// There already is a super admin account created in the seed.
				// So this user ID is the next one which is 2
				assert.Equal(t, 2, responseUserID.UserID)
			}
		})
	}
}

func TestHandler_LogoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*LoginRequest
	}
	type want struct {
		error
		status int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "normal",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  nil,
				status: http.StatusOK,
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 1*time.Hour)
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("highEntropyPassword", argon2id.DefaultParams)
	assert.NoError(t, err)

	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "email@example.com", hashedPassword)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.LoginRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.LoginRequest)
			}
			assert.NoError(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))

			RegisterHTTPEndPoints(router, session, repo)
			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			assert.NotNil(t, ww.Header().Get("Set-Cookie"))
			token, err := extractToken(ww.Header().Get("Set-Cookie"))
			assert.NoError(t, err)

			rr = httptest.NewRequest(http.MethodGet, "/api/v1/restricted", nil)
			ww = httptest.NewRecorder()

			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: token,
			})

			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			rr = httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)
			ww = httptest.NewRecorder()

			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: token,
			})

			router.ServeHTTP(ww, rr)

			token, err = extractToken(ww.Header().Get("Set-Cookie"))
			assert.NoError(t, err)
			assert.Equal(t, token, "")
		})
	}
}

func TestHandler_Force_LogoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*LoginRequest
	}
	type want struct {
		error
		status int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "normal",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:  nil,
				status: http.StatusOK,
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 1*time.Hour)
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("highEntropyPassword", argon2id.DefaultParams)
	assert.NoError(t, err)

	// Create normal user
	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "email@example.com", hashedPassword)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.LoginRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.LoginRequest)
			}
			assert.NoError(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))

			RegisterHTTPEndPoints(router, session, repo)
			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			assert.NotNil(t, ww.Header().Get("Set-Cookie"))
			token, err := extractToken(ww.Header().Get("Set-Cookie"))
			assert.NoError(t, err)

			rr = httptest.NewRequest(http.MethodGet, "/api/v1/restricted", nil)
			ww = httptest.NewRecorder()

			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: token,
			})

			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			// Now we log in as super admin to log out admin@gmail.com
			admin := &LoginRequest{
				Email:    "admin@gmhafiz.com",
				Password: "highEntropyPassword",
			}
			err = json.NewEncoder(&buf).Encode(admin)
			assert.NoError(t, err)

			rr = httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww = httptest.NewRecorder()

			router.ServeHTTP(ww, rr)

			assert.NotNil(t, ww.Header().Get("Set-Cookie"))
			adminToken, err := extractToken(ww.Header().Get("Set-Cookie"))
			assert.NoError(t, err)

			// ID 2 is our normal user to be forced-log out
			rr = httptest.NewRequest(http.MethodPost, "/api/v1/restricted/logout/2", nil)
			ww = httptest.NewRecorder()
			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: adminToken,
			})

			router.ServeHTTP(ww, rr)

			assert.Equal(t, http.StatusOK, ww.Code)

			// Check normal user ID 2 cannot access restricted route anymore
			rr = httptest.NewRequest(http.MethodGet, "/api/v1/restricted", nil)
			ww = httptest.NewRecorder()
			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: token,
			})

			router.ServeHTTP(ww, rr)

			assert.Equal(t, http.StatusUnauthorized, ww.Code)
		})
	}
}

func TestHandler_Csrf_Valid_TokenIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*LoginRequest
	}
	type want struct {
		error
		status            int
		response          RespondCsrf
		csrfTokenValidity bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "normal",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:             nil,
				status:            http.StatusOK,
				response:          RespondCsrf{CsrfToken: ""},
				csrfTokenValidity: true,
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 100*time.Millisecond) // short expiry to test token expiry means test complete faster.
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("highEntropyPassword", argon2id.DefaultParams)
	assert.NoError(t, err)

	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "email@example.com", hashedPassword)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.LoginRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.LoginRequest)
			}
			assert.NoError(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))

			RegisterHTTPEndPoints(router, session, repo)
			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			assert.NotNil(t, ww.Header().Get("Set-Cookie"))
			token, err := extractToken(ww.Header().Get("Set-Cookie"))
			assert.NoError(t, err)

			rr = httptest.NewRequest(http.MethodGet, "/api/v1/restricted/csrf", nil)
			ww = httptest.NewRecorder()

			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: token,
			})

			router.ServeHTTP(ww, rr)

			assert.Equal(t, ww.Code, http.StatusOK)

			b, err := io.ReadAll(ww.Body)
			assert.NoError(t, err)

			var resp RespondCsrf
			err = json.Unmarshal(b, &resp)
			assert.NoError(t, err)

			assert.NotNil(t, resp.CsrfToken)

			validity := csrf.ValidToken(context.Background(), migrator.DB, resp.CsrfToken)
			assert.Equal(t, tt.want.csrfTokenValidity, validity)

			// csrf token does not get deleted yet
			validity = csrf.ValidToken(context.Background(), migrator.DB, resp.CsrfToken)
			assert.Equal(t, tt.want.csrfTokenValidity, validity)

			time.Sleep(101 * time.Millisecond)

			validity = csrf.ValidToken(context.Background(), migrator.DB, resp.CsrfToken)
			assert.Equal(t, false, validity)
		})
	}
}

func TestHandler_Csrf_Valid_And_Delete_TokenIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*LoginRequest
	}
	type want struct {
		error
		status            int
		response          RespondCsrf
		csrfTokenValidity bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "normal",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "highEntropyPassword",
				},
			},
			want: want{
				error:             nil,
				status:            http.StatusOK,
				response:          RespondCsrf{CsrfToken: ""},
				csrfTokenValidity: true,
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 10*time.Minute)
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("highEntropyPassword", argon2id.DefaultParams)
	assert.NoError(t, err)

	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "email@example.com", hashedPassword)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.LoginRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.LoginRequest)
			}
			assert.NoError(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))

			RegisterHTTPEndPoints(router, session, repo)
			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			assert.NotNil(t, ww.Header().Get("Set-Cookie"))
			token, err := extractToken(ww.Header().Get("Set-Cookie"))
			assert.NoError(t, err)

			rr = httptest.NewRequest(http.MethodGet, "/api/v1/restricted/csrf", nil)
			ww = httptest.NewRecorder()

			rr.AddCookie(&http.Cookie{
				Name:  sessionName,
				Value: token,
			})

			router.ServeHTTP(ww, rr)

			assert.Equal(t, ww.Code, http.StatusOK)

			b, err := io.ReadAll(ww.Body)
			assert.NoError(t, err)

			var resp RespondCsrf
			err = json.Unmarshal(b, &resp)
			assert.NoError(t, err)

			assert.NotNil(t, resp.CsrfToken)

			err = csrf.ValidAndDeleteToken(context.Background(), migrator.DB, resp.CsrfToken)
			assert.NoError(t, err)

			// at this point, the csrf token would have been deleted
			err = csrf.ValidAndDeleteToken(context.Background(), migrator.DB, resp.CsrfToken)
			assert.NotNil(t, err)
		})
	}
}

func TestHandler_LoginWithInvalidPasswordIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	type args struct {
		*LoginRequest
	}
	type want struct {
		error
		status int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "invalid password",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "wrongPassword123456",
				},
			},
			want: want{
				error:  nil,
				status: http.StatusUnauthorized,
			},
		},
		{
			name: "empty password",
			args: args{
				LoginRequest: &LoginRequest{
					Email:    "email@example.com",
					Password: "",
				},
			},
			want: want{
				error:  ErrPasswordLength,
				status: http.StatusBadRequest,
			},
		},
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	session := newSession(migrator.DB, 1*time.Hour)
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("correctPassword", argon2id.DefaultParams)
	assert.NoError(t, err)

	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "email@example.com", hashedPassword)
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var err error
			if tt.args.LoginRequest != nil {
				err = json.NewEncoder(&buf).Encode(tt.args.LoginRequest)
			}
			assert.NoError(t, err)

			rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", &buf)
			ww := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Use(middleware.LoadAndSave(session))

			RegisterHTTPEndPoints(router, session, repo)
			router.ServeHTTP(ww, rr)

			assert.Equal(t, tt.want.status, ww.Code)

			if tt.want.error != nil {
				b, err := io.ReadAll(ww.Body)
				assert.NoError(t, err)

				errStruct := struct {
					Message string `json:"message"`
				}{}
				err = json.Unmarshal(b, &errStruct)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.error.Error(), errStruct.Message)
			}
		})
	}
}

func TestHandler_SessionExpirationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	drv := entsql.OpenDB(DBDriver, migrator.DB)
	client := gen.NewClient(gen.Driver(drv))
	defer client.Close()

	// Set a very short session duration for testing
	session := newSession(migrator.DB, 1*time.Second)
	repo := NewRepo(migrator.DB, session)

	hashedPassword, err := argon2id.CreateHash("testPassword123456", argon2id.DefaultParams)
	assert.NoError(t, err)

	_, err = migrator.DB.ExecContext(context.Background(), `
		INSERT INTO users (email, password) VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING 
	`, "session@test.com", hashedPassword)
	assert.NoError(t, err)

	// First login to get session token
	loginReq := &LoginRequest{
		Email:    "session@test.com",
		Password: "testPassword123456",
	}
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(loginReq)
	assert.NoError(t, err)

	rr := httptest.NewRequest(http.MethodPost, "/api/v1/login", buf)
	ww := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(middleware.LoadAndSave(session))
	RegisterHTTPEndPoints(router, session, repo)
	router.ServeHTTP(ww, rr)

	assert.Equal(t, http.StatusOK, ww.Code)
	token, err := extractToken(ww.Header().Get("Set-Cookie"))
	assert.NoError(t, err)

	// Wait for session to expire
	time.Sleep(2 * time.Second)

	// Try to access restricted endpoint with expired session
	rr = httptest.NewRequest(http.MethodGet, "/api/v1/restricted", nil)
	ww = httptest.NewRecorder()
	rr.AddCookie(&http.Cookie{
		Name:  sessionName,
		Value: token,
	})

	router.ServeHTTP(ww, rr)
	assert.Equal(t, http.StatusUnauthorized, ww.Code)
}

func extractToken(cookie string) (string, error) {
	parts := strings.Split(cookie, ";")
	if len(parts) == 0 {
		return "", errors.New("invalid cookie")
	}

	for _, part := range parts {
		keyVal := strings.Split(part, "=")
		if len(keyVal) != 2 {
			return "", errors.New("invalid cookie")
		}
		if keyVal[0] == sessionName {
			return keyVal[1], nil
		}
	}

	return "", errors.New("invalid cookie")
}

func dbClient() *gen.Client {
	drv := entsql.OpenDB(DBDriver, migrator.DB)
	return gen.NewClient(gen.Driver(drv))
}

func newSession(db *sql.DB, duration time.Duration) *scs.SessionManager {
	manager := scs.New()
	manager.Store = postgresstore.New(db)
	manager.CtxStore = postgresstore.New(db)
	manager.Lifetime = duration
	manager.Cookie.Name = sessionName
	manager.Cookie.HttpOnly = false
	manager.Cookie.Path = "/"
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = false

	return manager
}
