package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/gobank/internal/domain/entity"
	"github.com/yourusername/gobank/internal/domain/repository"
	"github.com/yourusername/gobank/internal/domain/service"
	"github.com/yourusername/gobank/internal/infrastructure/config"
	"github.com/yourusername/gobank/internal/pkg/apperror"
	"github.com/yourusername/gobank/internal/pkg/password"
	"github.com/yourusername/gobank/internal/pkg/token"
)

type userService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	passwordHasher   password.Hasher
	jwtManager       token.JWTManager
	config           *config.Config
}

func NewUserService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	passwordHasher password.Hasher,
	jwtManager token.JWTManager,
	cfg *config.Config,
) service.UserService {
	return &userService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordHasher:   passwordHasher,
		jwtManager:       jwtManager,
		config:           cfg,
	}
}

func (s *userService) Register(ctx context.Context, input *entity.CreateUserInput) (*entity.User, error) {
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to check email existence", 500)
	}
	if exists {
		return nil, apperror.ErrEmailAlreadyExists
	}

	hashedPassword, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to hash password", 500)
	}

	user := entity.NewUser(input.Email, hashedPassword, input.FullName)

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to create user", 500)
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, input *entity.LoginInput) (*entity.AuthTokens, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get user", 500)
	}
	if user == nil {
		return nil, apperror.ErrInvalidCredentials
	}

	if err := s.passwordHasher.Compare(user.PasswordHash, input.Password); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to generate access token", 500)
	}

	refreshToken, refreshTokenHash, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to generate refresh token", 500)
	}

	refreshTokenEntity := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(s.config.JWT.RefreshTokenExpiry),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to store refresh token", 500)
	}

	return &entity.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.JWT.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*entity.AuthTokens, error) {
	tokenHash := s.jwtManager.HashRefreshToken(refreshToken)

	storedToken, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to validate refresh token", 500)
	}
	if storedToken == nil {
		return nil, apperror.ErrInvalidToken
	}

	if storedToken.ExpiresAt.Before(time.Now()) {
		_ = s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
		return nil, apperror.ErrTokenExpired
	}

	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get user", 500)
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}

	if err := s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to delete old refresh token", 500)
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to generate access token", 500)
	}

	newRefreshToken, newRefreshTokenHash, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to generate refresh token", 500)
	}

	refreshTokenEntity := &entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: newRefreshTokenHash,
		ExpiresAt: time.Now().Add(s.config.JWT.RefreshTokenExpiry),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to store refresh token", 500)
	}

	return &entity.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.JWT.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := s.jwtManager.HashRefreshToken(refreshToken)
	return s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get user", 500)
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, input *entity.UpdateUserInput) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get user", 500)
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}

	if input.FullName != "" {
		user.FullName = input.FullName
	}

	if input.Email != "" && input.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
		if err != nil {
			return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to check email existence", 500)
		}
		if exists {
			return nil, apperror.ErrEmailAlreadyExists
		}
		user.Email = input.Email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to update user", 500)
	}

	return user, nil
}
