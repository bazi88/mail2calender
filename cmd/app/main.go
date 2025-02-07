// @title Go8 API
// @version 1.0
// @description This is a sample Go8 API server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	"mail2calendar/internal/config"
	"mail2calendar/internal/domain/health"
	"mail2calendar/internal/infrastructure/logger"
)

func main() {
	// Load config
	cfg := config.Load()

	// Setup logger
	log := logger.GetLogger()
	if cfg.API.RequestLog {
		log.SetLevel(logrus.DebugLevel)
	}

	// Setup database connection string
	dbURL := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.Driver,
		cfg.DB.User,
		cfg.DB.Pass,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	// Setup database
	db, err := sqlx.Connect(cfg.DB.Driver, dbURL)
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}

	// Configure database connection pool
	db.SetMaxOpenConns(cfg.DB.MaxConnectionPool)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.DB.ConnectionLifetime)

	// Setup Chi router
	router := chi.NewRouter()

	// Setup middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)

	// Setup CORS
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   strings.Split(cfg.CORS.AllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler)

	// Setup health check
	healthRepo := health.NewRepo(db)
	healthUseCase := health.New(healthRepo)
	health.RegisterHTTPEndPoints(router, healthUseCase)

	// Start server
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Infof("starting server on %s:%d", cfg.API.Host, cfg.API.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", err)
	}

	log.Info("server exited properly")
}
