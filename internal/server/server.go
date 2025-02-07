package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/gmhafiz/scs/v2"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/cors"

	"mail2calendar/config"
	"mail2calendar/ent/gen"
	"mail2calendar/internal/middleware"
	db "mail2calendar/third_party/database"
	"mail2calendar/third_party/postgresstore"
	"mail2calendar/third_party/validate"
)

type Server struct {
	Version string
	cfg     *config.Config

	db   *sql.DB
	sqlx *sqlx.DB
	ent  *gen.Client

	session       *scs.SessionManager
	sessionCloser *postgresstore.PostgresStore

	validator *validator.Validate
	cors      *cors.Cors
	router    *chi.Mux

	httpServer *http.Server
}

type Options func(opts *Server) error

func New(opts ...Options) *Server {
	s := defaultServer()

	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return s
}

func WithVersion(version string) Options {
	return func(opts *Server) error {
		log.Printf("Starting API version: %s\n", version)
		opts.Version = version
		return nil
	}
}

func defaultServer() *Server {
	return &Server{
		cfg:    config.New(),
		router: chi.NewRouter(),
	}
}

func (s *Server) Init() {
	s.setCors()
	s.NewDatabase()
	s.newValidator()
	s.newAuthentication()
	s.newRouter()
	s.setGlobalMiddleware()
	s.InitDomains()
}

func (s *Server) setCors() {
	s.cors = cors.New(
		cors.Options{
			AllowedOrigins: s.cfg.Cors.AllowedOrigins,
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		})
}

func (s *Server) NewDatabase() {
	if s.cfg.Database.Driver == "" {
		log.Fatal("please fill in database credentials in .env file or set in environment variable")
	}

	s.sqlx = db.NewSqlx(s.cfg.Database)
	s.sqlx.SetMaxOpenConns(s.cfg.Database.MaxConnectionPool)
	s.sqlx.SetMaxIdleConns(s.cfg.Database.MaxIdleConnections)
	s.sqlx.SetConnMaxLifetime(s.cfg.Database.ConnectionsMaxLifeTime)

	dsn := fmt.Sprintf("postgres://%s:%d/%s?sslmode=%s&user=%s&password=%s",
		s.cfg.Database.Host,
		s.cfg.Database.Port,
		s.cfg.Database.Name,
		s.cfg.Database.SslMode,
		s.cfg.Database.User,
		s.cfg.Database.Pass,
	)
	s.db = s.sqlx.DB
	s.newEnt(dsn)
}

func (s *Server) newValidator() {
	s.validator = validate.New()
}

func (s *Server) newAuthentication() {
	manager := scs.New()
	manager.Store = postgresstore.New(s.sqlx.DB)
	manager.CtxStore = postgresstore.New(s.sqlx.DB)
	manager.Lifetime = s.cfg.Session.Duration
	manager.Cookie.Name = s.cfg.Session.Name
	manager.Cookie.Domain = s.cfg.Session.Domain
	manager.Cookie.HttpOnly = s.cfg.Session.HttpOnly
	manager.Cookie.Path = s.cfg.Session.Path
	manager.Cookie.Persist = true
	manager.Cookie.SameSite = http.SameSite(s.cfg.Session.SameSite)
	manager.Cookie.Secure = s.cfg.Session.Secure

	s.sessionCloser = postgresstore.NewWithCleanupInterval(s.sqlx.DB, 30*time.Minute)

	s.session = manager
}

func (s *Server) newRouter() {
	s.router = chi.NewRouter()
}

func (s *Server) setGlobalMiddleware() {
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "endpoint not found"}`))
	})

	s.router.Use(s.cors.Handler)
	s.router.Use(middleware.Json)
	s.router.Use(middleware.LoadAndSave(s.session))
	if s.cfg.Api.RequestLog {
		s.router.Use(chiMiddleware.Logger)
	}
	s.router.Use(middleware.Recovery)
}

func (s *Server) newEnt(dsn string) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Println(err)
	}
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := gen.NewClient(gen.Driver(drv))

	s.ent = client
}

func (s *Server) Config() *config.Config {
	return s.cfg
}

func (s *Server) Run() {
	s.httpServer = &http.Server{
		Addr:              s.cfg.Api.Host + ":" + s.cfg.Api.Port,
		Handler:           s.router,
		ReadHeaderTimeout: s.cfg.Api.ReadHeaderTimeout,
	}

	go func() {
		log.Printf("Server is running on %s:%s\n", s.cfg.Api.Host, s.cfg.Api.Port)
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Api.GracefulTimeout*time.Second)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Println(err)
	}

	s.closeResources(ctx)
}

func (s *Server) closeResources(ctx context.Context) {
	_ = s.sqlx.Close()
	_ = s.ent.Close()
	s.sessionCloser.StopCleanup()
}
