package postgres

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/yourusername/gobank/internal/domain/entity"
	"github.com/yourusername/gobank/internal/domain/repository"
	"github.com/yourusername/gobank/internal/infrastructure/database"
)

type accountRepository struct {
	pool *pgxpool.Pool
}

func NewAccountRepository(db *database.PostgresDB) repository.AccountRepository {
	return &accountRepository{pool: db.Pool}
}

func (r *accountRepository) Create(ctx context.Context, account *entity.Account) error {
	if account.AccountNumber == "" {
		account.AccountNumber = generateAccountNumber()
	}

	query := `
		INSERT INTO accounts (id, user_id, account_number, account_type, currency, balance, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	if tx, ok := ctx.Value(database.TxKey{}).(pgx.Tx); ok {
		_, err := tx.Exec(ctx, query,
			account.ID,
			account.UserID,
			account.AccountNumber,
			account.AccountType,
			account.Currency,
			account.Balance,
			account.Status,
			account.CreatedAt,
			account.UpdatedAt,
		)
		return err
	}

	_, err := r.pool.Exec(ctx, query,
		account.ID,
		account.UserID,
		account.AccountNumber,
		account.AccountType,
		account.Currency,
		account.Balance,
		account.Status,
		account.CreatedAt,
		account.UpdatedAt,
	)
	return err
}

func (r *accountRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, currency, balance, status, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`
	account := &entity.Account{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.AccountType,
		&account.Currency,
		&account.Balance,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *accountRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, currency, balance, status, created_at, updated_at
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`

	account := &entity.Account{}
	var row pgx.Row

	if tx, ok := ctx.Value(database.TxKey{}).(pgx.Tx); ok {
		row = tx.QueryRow(ctx, query, id)
	} else {
		row = r.pool.QueryRow(ctx, query, id)
	}

	err := row.Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.AccountType,
		&account.Currency,
		&account.Balance,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *accountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*entity.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, currency, balance, status, created_at, updated_at
		FROM accounts
		WHERE account_number = $1
	`
	account := &entity.Account{}
	err := r.pool.QueryRow(ctx, query, accountNumber).Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNumber,
		&account.AccountType,
		&account.Currency,
		&account.Balance,
		&account.Status,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (r *accountRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, currency, balance, status, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*entity.Account
	for rows.Next() {
		account := &entity.Account{}
		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.AccountNumber,
			&account.AccountType,
			&account.Currency,
			&account.Balance,
			&account.Status,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

func (r *accountRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM accounts WHERE user_id = $1`
	var count int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *accountRepository) Update(ctx context.Context, account *entity.Account) error {
	query := `
		UPDATE accounts
		SET account_type = $2, currency = $3, status = $4, updated_at = NOW()
		WHERE id = $1
	`

	if tx, ok := ctx.Value(database.TxKey{}).(pgx.Tx); ok {
		_, err := tx.Exec(ctx, query,
			account.ID,
			account.AccountType,
			account.Currency,
			account.Status,
		)
		return err
	}

	_, err := r.pool.Exec(ctx, query,
		account.ID,
		account.AccountType,
		account.Currency,
		account.Status,
	)
	return err
}

func (r *accountRepository) UpdateBalance(ctx context.Context, id uuid.UUID, newBalance decimal.Decimal) error {
	query := `
		UPDATE accounts
		SET balance = $2, updated_at = NOW()
		WHERE id = $1
	`

	if tx, ok := ctx.Value(database.TxKey{}).(pgx.Tx); ok {
		_, err := tx.Exec(ctx, query, id, newBalance)
		return err
	}

	_, err := r.pool.Exec(ctx, query, id, newBalance)
	return err
}

func generateAccountNumber() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%010d", rng.Int63n(10000000000))
}
