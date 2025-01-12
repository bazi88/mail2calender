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
	"syscall"
	"time"

	_ "github.com/bazi88/mono-golang/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/bazi88/mono-golang/internal/config"
	"github.com/bazi88/mono-golang/internal/delivery/http/middleware"
	"github.com/bazi88/mono-golang/internal/domain/auth"
	"github.com/bazi88/mono-golang/internal/domain/author"
	"github.com/bazi88/mono-golang/internal/domain/book"
	"github.com/bazi88/mono-golang/internal/infrastructure/cache"
	"github.com/bazi88/mono-golang/internal/infrastructure/database"
	"github.com/bazi88/mono-golang/internal/infrastructure/logger"
	"github.com/bazi88/mono-golang/internal/infrastructure/metrics"
)

// @Summary Health check
// @Description Check if the API is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {string} string "OK"
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func main() {
	// Load config
	cfg := config.Load()

	// Setup logger
	logger := logger.New(cfg)

	// Setup metrics
	metrics := metrics.New(cfg)

	// Setup database
	db, err := database.New(cfg)
	if err != nil {
		logger.Fatal("failed to connect to database", err)
	}

	// Setup Redis cache
	cache, err := cache.New(cfg)
	if err != nil {
		logger.Fatal("failed to connect to Redis", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Setup CORS middleware
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", healthCheck)

	// Setup Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup API routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth.RegisterHTTPEndpoints(v1, db, cache, logger, metrics)

		// Author routes
		author.RegisterHTTPEndpoints(v1, db, cache, logger, metrics)

		// Book routes
		book.RegisterHTTPEndpoints(v1, db, cache, logger, metrics)
	}

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", err)
	}

	logger.Info("server exited properly")
}
