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

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"

	"mono-golang/internal/config"
	"mono-golang/internal/domain/ner/handler"
	"mono-golang/internal/domain/ner/usecase"
	"mono-golang/internal/grpc/client"
	"mono-golang/internal/middleware"
)

func main() {
	// Load config
	cfg := config.Load()

	// Khởi tạo Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Pass,
		DB:       cfg.Redis.Name,
	})

	// Khởi tạo NER client
	nerClient, err := client.NewNERClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create NER client: %v", err)
	}
	defer nerClient.Close()

	// Khởi tạo use case và handler
	nerUseCase := usecase.New(nerClient)

	// Khởi tạo router
	r := chi.NewRouter()

	// Khởi tạo rate limiter
	rateLimiter := middleware.DefaultRateLimiter(redisClient)

	// Đăng ký routes
	handler.Register(r, nerUseCase, rateLimiter)

	// Khởi tạo server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port),
		Handler: r,
	}

	// Channel để nhận signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Chạy server trong goroutine riêng
	go func() {
		log.Printf("Server is running on %s:%d", cfg.API.Host, cfg.API.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Đợi signal
	<-stop
	log.Println("Shutting down server...")

	// Tạo context với timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
