package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"live-poll-backend/internal/config"
	"live-poll-backend/internal/database"
	"live-poll-backend/internal/handlers"
	"live-poll-backend/internal/middleware"
	"live-poll-backend/internal/redis"
	"live-poll-backend/internal/repository"
	"live-poll-backend/internal/services"
	"live-poll-backend/internal/websocket"
)

func main() {
	// 1. Load application configuration
	cfg := config.LoadConfig()

	// 2. Set Gin execution mode and validate production configuration
	if cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
		if cfg.JWTSecret == "default_dev_secret_change_in_production" || len(cfg.JWTSecret) < 32 {
			log.Fatalf("[FATAL] Production configuration error: JWT_SECRET must be set to a cryptographically secure key of at least 32 characters in release mode")
		}
	}

	// 3. Establish storage connections with background contexts
	initCtx, initCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer initCancel()

	mongoDB, err := database.ConnectMongoDB(initCtx, cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		log.Printf("[WARN] MongoDB connection unverified on startup: %v", err)
	}

	redisClient, err := redis.ConnectRedis(initCtx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("[WARN] Redis connection unverified on startup: %v", err)
	}

	// 4. Initialize architectural layers (Repository -> Service -> Handler)
	healthRepo := repository.NewHealthRepository(mongoDB, redisClient)
	healthService := services.NewHealthService(healthRepo)
	healthHandler := handlers.NewHealthHandler(healthService)

	userRepo := repository.NewUserRepository(mongoDB)
	if mongoDB != nil && mongoDB.Database != nil {
		if err := userRepo.EnsureIndexes(initCtx); err != nil {
			log.Printf("[WARN] Failed to create user database indexes: %v", err)
		}
	}
	authService := services.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	authHandler := handlers.NewAuthHandler(authService)

	pollRepo := repository.NewPollRepository(mongoDB)
	if mongoDB != nil && mongoDB.Database != nil {
		if err := pollRepo.EnsureIndexes(initCtx); err != nil {
			log.Printf("[WARN] Failed to create poll database indexes: %v", err)
		}
	}
	wsHub := websocket.NewHub()
	go wsHub.Run()

	pubsubService := redis.NewPubSubService(redisClient)
	if err := pubsubService.Subscribe(context.Background(), func(pollID string, payload []byte) {
		wsHub.BroadcastToPoll(pollID, payload)
	}); err != nil {
		log.Printf("[WARN] Failed to start Redis Pub/Sub subscriber: %v", err)
	}

	voteCounter := redis.NewVoteCounter(redisClient)
	pollService := services.NewPollService(pollRepo, voteCounter, pubsubService)
	pollHandler := handlers.NewPollHandler(pollService)
	wsHandler := handlers.NewWebSocketHandler(wsHub, pollService)

	// 5. Initialize router and attach centralized middleware
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.MaxBodySizeMiddleware(2 * 1024 * 1024)) // 2MB max payload limit
	router.Use(middleware.CORSMiddleware(cfg.ClientOrigin))

	// Centralized 404 handler using standard response format
	router.NoRoute(func(c *gin.Context) {
		handlers.SendError(c, http.StatusNotFound, "NOT_FOUND", "The requested resource was not found")
	})

	// 6. Register REST API and WebSocket routes
	api := router.Group("/api")
	{
		api.GET("/health", healthHandler.GetHealth)

		// Public authentication endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
		}

		// Public poll, results, and real-time live WebSocket endpoints
		api.GET("/polls/:id", pollHandler.GetPoll)
		api.GET("/polls/:id/results", pollHandler.GetPollResults)
		api.POST("/polls/:id/vote", middleware.OptionalAuthMiddleware(authService), pollHandler.Vote)
		api.GET("/polls/:id/live", wsHandler.ServeWS)

		// Protected endpoints guarded by JWT authentication middleware
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.POST("/polls", pollHandler.CreatePoll)
			protected.GET("/my/polls", pollHandler.GetMyPolls)
			protected.DELETE("/polls/:id", pollHandler.DeletePoll)
			protected.PATCH("/polls/:id/status", pollHandler.SetPollStatus)
		}
	}

	// 7. Configure HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Start server in a background goroutine
	go func() {
		log.Printf("[INFO] LivePoll backend server listening on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] HTTP server error: %v", err)
		}
	}()

	// 9. Graceful shutdown listening for OS termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Termination signal received. Initiating graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] Server forced to shutdown: %v", err)
	} else {
		log.Println("[INFO] HTTP server stopped accepting connections")
	}

	// Close database connections
	if mongoDB != nil {
		if err := mongoDB.Disconnect(shutdownCtx); err != nil {
			log.Printf("[ERROR] Error disconnecting MongoDB: %v", err)
		}
	}

	if pubsubService != nil {
		if err := pubsubService.Close(); err != nil {
			log.Printf("[ERROR] Error closing Redis Pub/Sub: %v", err)
		}
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("[ERROR] Error closing Redis connection: %v", err)
		}
	}

	log.Println("[INFO] LivePoll backend successfully shutdown")
}
