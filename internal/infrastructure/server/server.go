package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourusername/gobank/internal/adapter/handler"
	"github.com/yourusername/gobank/internal/adapter/middleware"
	"github.com/yourusername/gobank/internal/adapter/repository/redis"
	"github.com/yourusername/gobank/internal/infrastructure/config"
	"github.com/yourusername/gobank/internal/infrastructure/logger"
	"github.com/yourusername/gobank/internal/pkg/token"
)

type Server struct {
	router          *gin.Engine
	httpServer      *http.Server
	config          *config.Config
	logger          *logger.Logger
	userHandler     *handler.UserHandler
	accountHandler  *handler.AccountHandler
	transferHandler *handler.TransferHandler
	healthHandler   *handler.HealthHandler
	jwtManager      token.JWTManager
	rateLimiter     *redis.RateLimiter
}

type ServerDeps struct {
	Config          *config.Config
	Logger          *logger.Logger
	UserHandler     *handler.UserHandler
	AccountHandler  *handler.AccountHandler
	TransferHandler *handler.TransferHandler
	HealthHandler   *handler.HealthHandler
	JWTManager      token.JWTManager
	RateLimiter     *redis.RateLimiter
}

func NewServer(deps *ServerDeps) *Server {
	if deps.Config.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	s := &Server{
		router:          router,
		config:          deps.Config,
		logger:          deps.Logger,
		userHandler:     deps.UserHandler,
		accountHandler:  deps.AccountHandler,
		transferHandler: deps.TransferHandler,
		healthHandler:   deps.HealthHandler,
		jwtManager:      deps.JWTManager,
		rateLimiter:     deps.RateLimiter,
	}

	s.setupMiddleware()
	s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         ":" + deps.Config.Server.Port,
		Handler:      router,
		ReadTimeout:  deps.Config.Server.ReadTimeout,
		WriteTimeout: deps.Config.Server.WriteTimeout,
	}

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Recovery(s.logger))
	s.router.Use(middleware.RequestID())
	s.router.Use(middleware.Logging(s.logger))
	s.router.Use(middleware.CORS())
	s.router.Use(middleware.SecurityHeaders())
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", s.healthHandler.Health)
	s.router.GET("/ready", s.healthHandler.Ready)
	s.router.GET("/info", s.healthHandler.Info)
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := s.router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.Use(middleware.RateLimitByIP(s.rateLimiter))
			auth.POST("/register", s.userHandler.Register)
			auth.POST("/login", s.userHandler.Login)
			auth.POST("/refresh", s.userHandler.RefreshToken)
			auth.POST("/logout", s.userHandler.Logout)
		}

		users := api.Group("/users")
		users.Use(middleware.Auth(s.jwtManager))
		users.Use(middleware.RateLimit(s.rateLimiter))
		{
			users.GET("/me", s.userHandler.GetMe)
			users.PUT("/me", s.userHandler.UpdateMe)
		}

		accounts := api.Group("/accounts")
		accounts.Use(middleware.Auth(s.jwtManager))
		accounts.Use(middleware.RateLimit(s.rateLimiter))
		{
			accounts.POST("", s.accountHandler.Create)
			accounts.GET("", s.accountHandler.List)
			accounts.GET("/:id", s.accountHandler.GetByID)
			accounts.GET("/:id/transactions", s.accountHandler.GetTransactions)
		}

		transfers := api.Group("/transfers")
		transfers.Use(middleware.Auth(s.jwtManager))
		transfers.Use(middleware.RateLimit(s.rateLimiter))
		{
			transfers.POST("", s.transferHandler.Create)
			transfers.GET("", s.transferHandler.List)
			transfers.GET("/:id", s.transferHandler.GetByID)
		}
	}
}

func (s *Server) Run() error {
	go func() {
		s.logger.Info().Str("port", s.config.Server.Port).Msg("Starting HTTP server")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.logger.Info().Msg("Server exited gracefully")
	return nil
}

func (s *Server) Router() *gin.Engine {
	return s.router
}
