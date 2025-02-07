package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mail2calendar/internal/config"
	"mail2calendar/internal/delivery/http/middleware"
	"mail2calendar/internal/domain/ner/handler"
	"mail2calendar/internal/domain/ner/usecase"
	"mail2calendar/internal/grpc/client"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Pass,
		DB:       cfg.Redis.Name,
	})

	// Initialize router
	r := chi.NewRouter()

	// Add middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize rate limiter
	rateLimiter := middleware.NewRedisRateLimiter(redisClient, 10, time.Minute)
	r.Use(rateLimiter.Limit)

	// Initialize NER client
	nerClient, err := client.NewNERClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create NER client: %v", err)
	}
	defer nerClient.Close()

	// Initialize use case and handler
	nerUseCase := usecase.New(nerClient)

	// Register routes
	handler.RegisterRoutes(r, nerUseCase, rateLimiter)

	// Initialize server
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	// Channel for signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run server in goroutine
	go func() {
		log.Printf("Server is running on %s:%d", cfg.API.Host, cfg.API.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for signal
	<-stop
	log.Println("Shutting down server...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
