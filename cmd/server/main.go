package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/leconcepteur/cloud-cicd-poc/internal/api"
	"github.com/leconcepteur/cloud-cicd-poc/internal/auth"
	"github.com/leconcepteur/cloud-cicd-poc/internal/config"
	"github.com/leconcepteur/cloud-cicd-poc/internal/database"
	"github.com/leconcepteur/cloud-cicd-poc/internal/game"
	"github.com/leconcepteur/cloud-cicd-poc/internal/matchmaking"
	"github.com/leconcepteur/cloud-cicd-poc/internal/middleware"
)

func main() {
	cfg := config.Load()

	// Initialize databases
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		logFatal("Failed to connect to PostgreSQL", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		logFatal("Failed to run migrations", err)
	}

	redis, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		logFatal("Failed to connect to Redis", err)
	}
	defer redis.Close()

	// Initialize services
	sessionManager := auth.NewSessionManager(redis, cfg.SessionMaxAge)
	authService := auth.NewService(db, sessionManager)
	gameService := game.NewService(db)
	eventHub := matchmaking.NewEventHub()
	queue := matchmaking.NewQueue(redis)
	matchService := matchmaking.NewService(queue, gameService, eventHub)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService, cfg.SessionMaxAge)
	gameHandler := api.NewGameHandler(gameService, eventHub)
	matchHandler := api.NewMatchmakingHandler(matchService)
	eventsHandler := api.NewEventsHandler(eventHub)
	statsHandler := api.NewStatsHandler(authService)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	// Setup Echo
	e := echo.New()
	e.HideBanner = true

	// Global middleware
	e.Use(middleware.JSONLoggerMiddleware())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))
	e.Use(middleware.RateLimitMiddleware(rateLimiter))

	// Health check
	e.GET("/health", eventsHandler.Health)

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Auth routes (public)
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/logout", authHandler.Logout)

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(authService))

	// Auth protected routes
	protected.GET("/auth/me", authHandler.Me)

	// Matchmaking routes
	protected.POST("/matchmaking/join", matchHandler.Join)
	protected.POST("/matchmaking/leave", matchHandler.Leave)

	// Game routes
	protected.POST("/game/ready", gameHandler.Ready)
	protected.POST("/game/move", gameHandler.Move)
	protected.POST("/game/forfeit", gameHandler.Forfeit)
	protected.GET("/game/state", gameHandler.GetState)

	// Stats routes
	protected.GET("/stats", statsHandler.GetStats)

	// SSE events route
	protected.GET("/events", eventsHandler.Stream)

	// Start matchmaking service
	ctx, cancel := context.WithCancel(context.Background())
	go matchService.StartMatchmaking(ctx)

	// Start server
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		logInfo("Starting server", map[string]string{"port": cfg.Port, "env": cfg.Env})
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logFatal("Server failed", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logInfo("Shutting down server...", nil)
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logFatal("Server forced to shutdown", err)
	}

	logInfo("Server exited", nil)
}

func logInfo(message string, fields map[string]string) {
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     "info",
		"message":   message,
	}
	for k, v := range fields {
		entry[k] = v
	}
	jsonBytes, _ := json.Marshal(entry)
	fmt.Println(string(jsonBytes))
}

func logFatal(message string, err error) {
	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     "fatal",
		"message":   message,
		"error":     err.Error(),
	}
	jsonBytes, _ := json.Marshal(entry)
	fmt.Println(string(jsonBytes))
	os.Exit(1)
}
