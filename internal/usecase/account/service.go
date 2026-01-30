package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/gobank/internal/domain/entity"
	"github.com/yourusername/gobank/internal/domain/repository"
	"github.com/yourusername/gobank/internal/domain/service"
	"github.com/yourusername/gobank/internal/pkg/apperror"
)

type accountService struct {
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
}

func NewAccountService(
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
) service.AccountService {
	return &accountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *accountService) Create(ctx context.Context, userID uuid.UUID, input *entity.CreateAccountInput) (*entity.Account, error) {
	account := entity.NewAccount(userID, "", input.AccountType, input.Currency)

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to create account", 500)
	}

	createdAccount, err := s.accountRepo.GetByID(ctx, account.ID)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get created account", 500)
	}

	return createdAccount, nil
}

func (s *accountService) GetByID(ctx context.Context, userID, accountID uuid.UUID) (*entity.Account, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get account", 500)
	}
	if account == nil {
		return nil, apperror.ErrAccountNotFound
	}

	if account.UserID != userID {
		return nil, apperror.ErrForbidden
	}

	return account, nil
}

func (s *accountService) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entity.Account, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	accounts, err := s.accountRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get accounts", 500)
	}

	total, err := s.accountRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to count accounts", 500)
	}

	return accounts, total, nil
}

func (s *accountService) GetTransactions(ctx context.Context, userID, accountID uuid.UUID, page, pageSize int) ([]*entity.Transaction, int64, error) {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, 0, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get account", 500)
	}
	if account == nil {
		return nil, 0, apperror.ErrAccountNotFound
	}

	if account.UserID != userID {
		return nil, 0, apperror.ErrForbidden
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	transactions, err := s.transactionRepo.GetByAccountID(ctx, accountID, pageSize, offset)
	if err != nil {
		return nil, 0, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get transactions", 500)
	}

	total, err := s.transactionRepo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, 0, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to count transactions", 500)
	}

	return transactions, total, nil
}
