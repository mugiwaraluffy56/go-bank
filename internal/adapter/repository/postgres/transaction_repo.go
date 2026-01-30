package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/gobank/internal/domain/entity"
	"github.com/yourusername/gobank/internal/domain/repository"
	"github.com/yourusername/gobank/internal/infrastructure/database"
)

type transactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(db *database.PostgresDB) repository.TransactionRepository {
	return &transactionRepository{pool: db.Pool}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	query := `
		INSERT INTO transactions (id, account_id, type, amount, balance_after, description, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	if tx, ok := ctx.Value(database.TxKey{}).(pgx.Tx); ok {
		_, err := tx.Exec(ctx, query,
			transaction.ID,
			transaction.AccountID,
			transaction.Type,
			transaction.Amount,
			transaction.BalanceAfter,
			transaction.Description,
			transaction.ReferenceID,
			transaction.CreatedAt,
		)
		return err
	}

	_, err := r.pool.Exec(ctx, query,
		transaction.ID,
		transaction.AccountID,
		transaction.Type,
		transaction.Amount,
		transaction.BalanceAfter,
		transaction.Description,
		transaction.ReferenceID,
		transaction.CreatedAt,
	)
	return err
}

func (r *transactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	query := `
		SELECT id, account_id, type, amount, balance_after, description, reference_id, created_at
		FROM transactions
		WHERE id = $1
	`
	tx := &entity.Transaction{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tx.ID,
		&tx.AccountID,
		&tx.Type,
		&tx.Amount,
		&tx.BalanceAfter,
		&tx.Description,
		&tx.ReferenceID,
		&tx.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *transactionRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entity.Transaction, error) {
	query := `
		SELECT id, account_id, type, amount, balance_after, description, reference_id, created_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*entity.Transaction
	for rows.Next() {
		tx := &entity.Transaction{}
		if err := rows.Scan(
			&tx.ID,
			&tx.AccountID,
			&tx.Type,
			&tx.Amount,
			&tx.BalanceAfter,
			&tx.Description,
			&tx.ReferenceID,
			&tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, rows.Err()
}

func (r *transactionRepository) GetByAccountIDAndDateRange(ctx context.Context, accountID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]*entity.Transaction, error) {
	query := `
		SELECT id, account_id, type, amount, balance_after, description, reference_id, created_at
		FROM transactions
		WHERE account_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.pool.Query(ctx, query, accountID, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*entity.Transaction
	for rows.Next() {
		tx := &entity.Transaction{}
		if err := rows.Scan(
			&tx.ID,
			&tx.AccountID,
			&tx.Type,
			&tx.Amount,
			&tx.BalanceAfter,
			&tx.Description,
			&tx.ReferenceID,
			&tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, rows.Err()
}

func (r *transactionRepository) CountByAccountID(ctx context.Context, accountID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE account_id = $1`
	var count int64
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&count)
	return count, err
}

type transferRepository struct {
	pool *pgxpool.Pool
}

func NewTransferRepository(db *database.PostgresDB) repository.TransferRepository {
	return &transferRepository{pool: db.Pool}
}

func (r *transferRepository) Create(ctx context.Context, transfer *entity.Transfer) error {
	query := `
		INSERT INTO transfers (id, idempotency_key, from_account_id, to_account_id, amount, currency, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	if tx, ok := ctx.Value(database.TxKey{}).(pgx.Tx); ok {
		_, err := tx.Exec(ctx, query,
			transfer.ID,
			transfer.IdempotencyKey,
			transfer.FromAccountID,
			transfer.ToAccountID,
			transfer.Amount,
			transfer.Currency,
			transfer.Status,
			transfer.CreatedAt,
		)
		return err
	}

	_, err := r.pool.Exec(ctx, query,
		transfer.ID,
		transfer.IdempotencyKey,
		transfer.FromAccountID,
		transfer.ToAccountID,
		transfer.Amount,
		transfer.Currency,
		transfer.Status,
		transfer.CreatedAt,
	)
	return err
}

func (r *transferRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transfer, error) {
	query := `
		SELECT id, idempotency_key, from_account_id, to_account_id, amount, currency, status, created_at, completed_at
		FROM transfers
		WHERE id = $1
	`
	transfer := &entity.Transfer{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&transfer.ID,
		&transfer.IdempotencyKey,
		&transfer.FromAccountID,
		&transfer.ToAccountID,
		&transfer.Amount,
		&transfer.Currency,
		&transfer.Status,
		&transfer.CreatedAt,
		&transfer.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return transfer, nil
}

func (r *transferRepository) GetByIdempotencyKey(ctx context.Context, key string) (*entity.Transfer, error) {
	query := `
		SELECT id, idempotency_key, from_account_id, to_account_id, amount, currency, status, created_at, completed_at
		FROM transfers
		WHERE idempotency_key = $1
	`
	transfer := &entity.Transfer{}
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&transfer.ID,
		&transfer.IdempotencyKey,
		&transfer.FromAccountID,
		&transfer.ToAccountID,
		&transfer.Amount,
		&transfer.Currency,
		&transfer.Status,
		&transfer.CreatedAt,
		&transfer.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return transfer, nil
}

func (r *transferRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Transfer, error) {
	query := `
		SELECT DISTINCT t.id, t.idempotency_key, t.from_account_id, t.to_account_id, t.amount, t.currency, t.status, t.created_at, t.completed_at
		FROM transfers t
		JOIN accounts a ON (t.from_account_id = a.id OR t.to_account_id = a.id)
		WHERE a.user_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []*entity.Transfer
	for rows.Next() {
		transfer := &entity.Transfer{}
		if err := rows.Scan(
			&transfer.ID,
			&transfer.IdempotencyKey,
			&transfer.FromAccountID,
			&transfer.ToAccountID,
			&transfer.Amount,
			&transfer.Currency,
			&transfer.Status,
			&transfer.CreatedAt,
			&transfer.CompletedAt,
		); err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	return transfers, rows.Err()
}

func (r *transferRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TransferStatus, completedAt *time.Time) error {
	query := `
		UPDATE transfers
		SET status = $2, completed_at = $3
		WHERE id = $1
	`

	if tx, ok := ctx.Value(database.TxKey{}).(pgx.Tx); ok {
		_, err := tx.Exec(ctx, query, id, status, completedAt)
		return err
	}

	_, err := r.pool.Exec(ctx, query, id, status, completedAt)
	return err
}

type auditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(db *database.PostgresDB) repository.AuditLogRepository {
	return &auditLogRepository{pool: db.Pool}
}

func (r *auditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		log.ID,
		log.UserID,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.OldValues,
		log.NewValues,
		log.IPAddress,
		log.UserAgent,
		log.CreatedAt,
	)
	return err
}

func (r *auditLogRepository) GetByEntityID(ctx context.Context, entityType string, entityID uuid.UUID, limit, offset int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, entityType, entityID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{}
		if err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.OldValues,
			&log.NewValues,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *auditLogRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*entity.AuditLog
	for rows.Next() {
		log := &entity.AuditLog{}
		if err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.EntityType,
			&log.EntityID,
			&log.OldValues,
			&log.NewValues,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
