package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountType string
type AccountStatus string
type Currency string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"

	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusFrozen   AccountStatus = "frozen"

	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
)

type Account struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	AccountNumber string          `json:"account_number"`
	AccountType   AccountType     `json:"account_type"`
	Currency      Currency        `json:"currency"`
	Balance       decimal.Decimal `json:"balance"`
	Status        AccountStatus   `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type CreateAccountInput struct {
	AccountType AccountType `json:"account_type" validate:"required,oneof=checking savings"`
	Currency    Currency    `json:"currency" validate:"required,oneof=USD EUR GBP"`
}

type AccountResponse struct {
	ID            uuid.UUID       `json:"id"`
	AccountNumber string          `json:"account_number"`
	AccountType   AccountType     `json:"account_type"`
	Currency      Currency        `json:"currency"`
	Balance       string          `json:"balance"`
	Status        AccountStatus   `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
}

func NewAccount(userID uuid.UUID, accountNumber string, accountType AccountType, currency Currency) *Account {
	now := time.Now().UTC()
	return &Account{
		ID:            uuid.New(),
		UserID:        userID,
		AccountNumber: accountNumber,
		AccountType:   accountType,
		Currency:      currency,
		Balance:       decimal.Zero,
		Status:        AccountStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (a *Account) ToResponse() *AccountResponse {
	return &AccountResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		AccountType:   a.AccountType,
		Currency:      a.Currency,
		Balance:       a.Balance.StringFixed(2),
		Status:        a.Status,
		CreatedAt:     a.CreatedAt,
	}
}

func (a *Account) CanDebit(amount decimal.Decimal) bool {
	return a.Status == AccountStatusActive && a.Balance.GreaterThanOrEqual(amount)
}

func (a *Account) CanCredit() bool {
	return a.Status == AccountStatusActive
}
