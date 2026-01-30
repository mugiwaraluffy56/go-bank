package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string
type TransferStatus string

const (
	TransactionTypeCredit TransactionType = "credit"
	TransactionTypeDebit  TransactionType = "debit"

	TransferStatusPending   TransferStatus = "pending"
	TransferStatusCompleted TransferStatus = "completed"
	TransferStatusFailed    TransferStatus = "failed"
)

type Transaction struct {
	ID           uuid.UUID       `json:"id"`
	AccountID    uuid.UUID       `json:"account_id"`
	Type         TransactionType `json:"type"`
	Amount       decimal.Decimal `json:"amount"`
	BalanceAfter decimal.Decimal `json:"balance_after"`
	Description  string          `json:"description"`
	ReferenceID  *uuid.UUID      `json:"reference_id,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type Transfer struct {
	ID             uuid.UUID       `json:"id"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	FromAccountID  uuid.UUID       `json:"from_account_id"`
	ToAccountID    uuid.UUID       `json:"to_account_id"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       Currency        `json:"currency"`
	Status         TransferStatus  `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}

type CreateTransferInput struct {
	FromAccountID  uuid.UUID `json:"from_account_id" validate:"required"`
	ToAccountID    uuid.UUID `json:"to_account_id" validate:"required,nefield=FromAccountID"`
	Amount         string    `json:"amount" validate:"required"`
	IdempotencyKey string    `json:"idempotency_key" validate:"omitempty,max=255"`
}

type TransferResponse struct {
	ID             uuid.UUID      `json:"id"`
	FromAccountID  uuid.UUID      `json:"from_account_id"`
	ToAccountID    uuid.UUID      `json:"to_account_id"`
	Amount         string         `json:"amount"`
	Currency       Currency       `json:"currency"`
	Status         TransferStatus `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
}

type TransactionResponse struct {
	ID           uuid.UUID       `json:"id"`
	Type         TransactionType `json:"type"`
	Amount       string          `json:"amount"`
	BalanceAfter string          `json:"balance_after"`
	Description  string          `json:"description"`
	CreatedAt    time.Time       `json:"created_at"`
}

type AuditLog struct {
	ID         uuid.UUID              `json:"id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
	OldValues  map[string]interface{} `json:"old_values,omitempty"`
	NewValues  map[string]interface{} `json:"new_values,omitempty"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	CreatedAt  time.Time              `json:"created_at"`
}

func NewTransfer(fromAccountID, toAccountID uuid.UUID, amount decimal.Decimal, currency Currency, idempotencyKey *string) *Transfer {
	return &Transfer{
		ID:             uuid.New(),
		IdempotencyKey: idempotencyKey,
		FromAccountID:  fromAccountID,
		ToAccountID:    toAccountID,
		Amount:         amount,
		Currency:       currency,
		Status:         TransferStatusPending,
		CreatedAt:      time.Now().UTC(),
	}
}

func NewTransaction(accountID uuid.UUID, txType TransactionType, amount, balanceAfter decimal.Decimal, description string, referenceID *uuid.UUID) *Transaction {
	return &Transaction{
		ID:           uuid.New(),
		AccountID:    accountID,
		Type:         txType,
		Amount:       amount,
		BalanceAfter: balanceAfter,
		Description:  description,
		ReferenceID:  referenceID,
		CreatedAt:    time.Now().UTC(),
	}
}

func (t *Transfer) ToResponse() *TransferResponse {
	return &TransferResponse{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		Amount:        t.Amount.StringFixed(2),
		Currency:      t.Currency,
		Status:        t.Status,
		CreatedAt:     t.CreatedAt,
		CompletedAt:   t.CompletedAt,
	}
}

func (t *Transaction) ToResponse() *TransactionResponse {
	return &TransactionResponse{
		ID:           t.ID,
		Type:         t.Type,
		Amount:       t.Amount.StringFixed(2),
		BalanceAfter: t.BalanceAfter.StringFixed(2),
		Description:  t.Description,
		CreatedAt:    t.CreatedAt,
	}
}
