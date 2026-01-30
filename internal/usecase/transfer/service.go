package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yourusername/gobank/internal/domain/entity"
	"github.com/yourusername/gobank/internal/domain/repository"
	"github.com/yourusername/gobank/internal/domain/service"
	"github.com/yourusername/gobank/internal/infrastructure/database"
	"github.com/yourusername/gobank/internal/pkg/apperror"
)

type transferService struct {
	accountRepo     repository.AccountRepository
	transferRepo    repository.TransferRepository
	transactionRepo repository.TransactionRepository
	db              *database.PostgresDB
}

func NewTransferService(
	accountRepo repository.AccountRepository,
	transferRepo repository.TransferRepository,
	transactionRepo repository.TransactionRepository,
	db *database.PostgresDB,
) service.TransferService {
	return &transferService{
		accountRepo:     accountRepo,
		transferRepo:    transferRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

func (s *transferService) Create(ctx context.Context, userID uuid.UUID, input *entity.CreateTransferInput) (*entity.Transfer, error) {
	if input.IdempotencyKey != "" {
		existingTransfer, err := s.transferRepo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
		if err != nil {
			return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to check idempotency key", 500)
		}
		if existingTransfer != nil {
			return existingTransfer, nil
		}
	}

	amount, err := decimal.NewFromString(input.Amount)
	if err != nil {
		return nil, apperror.ErrInvalidAmount
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperror.ErrInvalidAmount
	}

	if input.FromAccountID == input.ToAccountID {
		return nil, apperror.ErrSameAccount
	}

	var transfer *entity.Transfer

	err = s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		fromAccount, err := s.accountRepo.GetByIDForUpdate(txCtx, input.FromAccountID)
		if err != nil {
			return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get source account", 500)
		}
		if fromAccount == nil {
			return apperror.ErrAccountNotFound
		}

		if fromAccount.UserID != userID {
			return apperror.ErrForbidden
		}

		toAccount, err := s.accountRepo.GetByIDForUpdate(txCtx, input.ToAccountID)
		if err != nil {
			return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get destination account", 500)
		}
		if toAccount == nil {
			return apperror.ErrAccountNotFound
		}

		if fromAccount.Currency != toAccount.Currency {
			return apperror.ErrCurrencyMismatch
		}

		if !fromAccount.CanDebit(amount) {
			return apperror.ErrInsufficientBalance
		}

		if !toAccount.CanCredit() {
			return apperror.ErrAccountInactive
		}

		var idempotencyKey *string
		if input.IdempotencyKey != "" {
			idempotencyKey = &input.IdempotencyKey
		}

		transfer = entity.NewTransfer(
			input.FromAccountID,
			input.ToAccountID,
			amount,
			fromAccount.Currency,
			idempotencyKey,
		)

		if err := s.transferRepo.Create(txCtx, transfer); err != nil {
			return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to create transfer", 500)
		}

		newFromBalance := fromAccount.Balance.Sub(amount)
		if err := s.accountRepo.UpdateBalance(txCtx, fromAccount.ID, newFromBalance); err != nil {
			return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to update source account balance", 500)
		}

		newToBalance := toAccount.Balance.Add(amount)
		if err := s.accountRepo.UpdateBalance(txCtx, toAccount.ID, newToBalance); err != nil {
			return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to update destination account balance", 500)
		}

		debitTx := entity.NewTransaction(
			fromAccount.ID,
			entity.TransactionTypeDebit,
			amount,
			newFromBalance,
			fmt.Sprintf("Transfer to account %s", toAccount.AccountNumber),
			&transfer.ID,
		)
		if err := s.transactionRepo.Create(txCtx, debitTx); err != nil {
			return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to create debit transaction", 500)
		}

		creditTx := entity.NewTransaction(
			toAccount.ID,
			entity.TransactionTypeCredit,
			amount,
			newToBalance,
			fmt.Sprintf("Transfer from account %s", fromAccount.AccountNumber),
			&transfer.ID,
		)
		if err := s.transactionRepo.Create(txCtx, creditTx); err != nil {
			return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to create credit transaction", 500)
		}

		completedAt := time.Now().UTC()
		if err := s.transferRepo.UpdateStatus(txCtx, transfer.ID, entity.TransferStatusCompleted, &completedAt); err != nil {
			return apperror.Wrap(err, "INTERNAL_ERROR", "Failed to update transfer status", 500)
		}
		transfer.Status = entity.TransferStatusCompleted
		transfer.CompletedAt = &completedAt

		return nil
	})

	if err != nil {
		return nil, err
	}

	return transfer, nil
}

func (s *transferService) GetByID(ctx context.Context, userID uuid.UUID, transferID uuid.UUID) (*entity.Transfer, error) {
	transfer, err := s.transferRepo.GetByID(ctx, transferID)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get transfer", 500)
	}
	if transfer == nil {
		return nil, apperror.ErrTransferNotFound
	}

	fromAccount, err := s.accountRepo.GetByID(ctx, transfer.FromAccountID)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get account", 500)
	}

	toAccount, err := s.accountRepo.GetByID(ctx, transfer.ToAccountID)
	if err != nil {
		return nil, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get account", 500)
	}

	if (fromAccount != nil && fromAccount.UserID == userID) || (toAccount != nil && toAccount.UserID == userID) {
		return transfer, nil
	}

	return nil, apperror.ErrForbidden
}

func (s *transferService) GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entity.Transfer, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	transfers, err := s.transferRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, apperror.Wrap(err, "INTERNAL_ERROR", "Failed to get transfers", 500)
	}

	return transfers, int64(len(transfers)), nil
}
