package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/gobank/internal/domain/entity"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entity.Transaction, error)
	GetByAccountIDAndDateRange(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]*entity.Transaction, error)
	CountByAccountID(ctx context.Context, accountID uuid.UUID) (int64, error)
}

type TransferRepository interface {
	Create(ctx context.Context, transfer *entity.Transfer) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transfer, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*entity.Transfer, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Transfer, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TransferStatus, completedAt *time.Time) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	GetByEntityID(ctx context.Context, entityType string, entityID uuid.UUID, limit, offset int) ([]*entity.AuditLog, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.AuditLog, error)
}

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
