package main

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/gobank/internal/adapter/handler"
	"github.com/yourusername/gobank/internal/adapter/repository/postgres"
	redisRepo "github.com/yourusername/gobank/internal/adapter/repository/redis"
	"github.com/yourusername/gobank/internal/infrastructure/config"
	"github.com/yourusername/gobank/internal/infrastructure/database"
	"github.com/yourusername/gobank/internal/infrastructure/logger"
	"github.com/yourusername/gobank/internal/infrastructure/server"
	"github.com/yourusername/gobank/internal/pkg/password"
	"github.com/yourusername/gobank/internal/pkg/token"
	"github.com/yourusername/gobank/internal/pkg/validator"
	accountUsecase "github.com/yourusername/gobank/internal/usecase/account"
	transferUsecase "github.com/yourusername/gobank/internal/usecase/transfer"
	userUsecase "github.com/yourusername/gobank/internal/usecase/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger := logger.New(cfg.Server.Environment)
	appLogger.Info().Str("environment", cfg.Server.Environment).Msg("Starting GoBank API")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPostgresDB(ctx, &cfg.Database)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()
	appLogger.Info().Msg("Connected to PostgreSQL")

	redisDB, err := database.NewRedisDB(ctx, &cfg.Redis)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisDB.Close()
	appLogger.Info().Msg("Connected to Redis")

	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	accountRepo := postgres.NewAccountRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	transferRepo := postgres.NewTransferRepository(db)

	passwordHasher := password.NewHasher()

	jwtManager := token.NewJWTManager(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
		cfg.JWT.Issuer,
	)

	validatorInstance := validator.New()

	rateLimiter := redisRepo.NewRateLimiter(redisDB, cfg.RateLimit.RequestsPerMinute)

	userService := userUsecase.NewUserService(
		userRepo,
		refreshTokenRepo,
		passwordHasher,
		jwtManager,
		cfg,
	)

	accountService := accountUsecase.NewAccountService(
		accountRepo,
		transactionRepo,
	)

	transferService := transferUsecase.NewTransferService(
		accountRepo,
		transferRepo,
		transactionRepo,
		db,
	)

	userHandler := handler.NewUserHandler(userService, validatorInstance)
	accountHandler := handler.NewAccountHandler(accountService, validatorInstance)
	transferHandler := handler.NewTransferHandler(transferService, validatorInstance)
	healthHandler := handler.NewHealthHandler(db, redisDB)

	srv := server.NewServer(&server.ServerDeps{
		Config:          cfg,
		Logger:          appLogger,
		UserHandler:     userHandler,
		AccountHandler:  accountHandler,
		TransferHandler: transferHandler,
		HealthHandler:   healthHandler,
		JWTManager:      jwtManager,
		RateLimiter:     rateLimiter,
	})

	if err := srv.Run(); err != nil {
		appLogger.Fatal().Err(err).Msg("Server error")
	}
}
