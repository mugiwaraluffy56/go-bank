package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/gobank/internal/domain/entity"
)

type UserService interface {
	Register(ctx context.Context, input *entity.CreateUserInput) (*entity.User, error)
	Login(ctx context.Context, input *entity.LoginInput) (*entity.AuthTokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entity.AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Update(ctx context.Context, id uuid.UUID, input *entity.UpdateUserInput) (*entity.User, error)
}

type AccountService interface {
	Create(ctx context.Context, userID uuid.UUID, input *entity.CreateAccountInput) (*entity.Account, error)
	GetByID(ctx context.Context, userID, accountID uuid.UUID) (*entity.Account, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entity.Account, int64, error)
	GetTransactions(ctx context.Context, userID, accountID uuid.UUID, page, pageSize int) ([]*entity.Transaction, int64, error)
}

type TransferService interface {
	Create(ctx context.Context, userID uuid.UUID, input *entity.CreateTransferInput) (*entity.Transfer, error)
	GetByID(ctx context.Context, userID uuid.UUID, transferID uuid.UUID) (*entity.Transfer, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entity.Transfer, int64, error)
}

type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
