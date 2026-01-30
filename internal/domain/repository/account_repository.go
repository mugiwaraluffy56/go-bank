package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yourusername/gobank/internal/domain/entity"
)

type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*entity.Account, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Account, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	Update(ctx context.Context, account *entity.Account) error
	UpdateBalance(ctx context.Context, id uuid.UUID, newBalance decimal.Decimal) error
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*entity.Account, error)
}
