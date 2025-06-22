package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/signaling-server/internal/config"
	"github.com/signaling-server/internal/handler"
	"github.com/signaling-server/internal/middleware"
	"github.com/signaling-server/internal/repository"
	"github.com/signaling-server/internal/service"
	"github.com/signaling-server/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.New()
	log.Info("Starting signaling server...")

	// Load configuration
	cfg := config.Load()
	log.Infof("Server configuration loaded: %s:%s", cfg.Server.Host, cfg.Server.Port)

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Errorf("Failed to connect to Redis: %v", err)
		os.Exit(1)
	}
	log.Info("Connected to Redis successfully")

	// Initialize repositories
	redisRepo := repository.NewRedisRepository(redisClient)

	// Initialize services
	userService := service.NewUserService(redisRepo)
	roomService := service.NewRoomService(redisRepo, redisRepo)
	signalingService := service.NewSignalingService(userService, roomService, redisRepo, log)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	wsHandler := handler.NewWebSocketHandler(signalingService, userService, cfg, log)

	// Setup HTTP server with middleware
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)

	// WebSocket endpoint with middleware
	wsEndpoint := middleware.SessionMiddleware(http.HandlerFunc(wsHandler.HandleWebSocket))
	mux.Handle("/ws", middleware.CORSMiddleware(wsEndpoint))

	// Static file serving for development/testing
	mux.Handle("/", http.FileServer(http.Dir("./web/static/")))

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Server failed to start: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Errorf("Failed to close Redis connection: %v", err)
	}

	log.Info("Server exited")
}
