package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common errors
var (
	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}

	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Unauthorized access",
		StatusCode: http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "Access forbidden",
		StatusCode: http.StatusForbidden,
	}

	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "Invalid request",
		StatusCode: http.StatusBadRequest,
	}

	ErrInternalServer = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Internal server error",
		StatusCode: http.StatusInternalServerError,
	}

	ErrConflict = &AppError{
		Code:       "CONFLICT",
		Message:    "Resource conflict",
		StatusCode: http.StatusConflict,
	}

	ErrValidation = &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		StatusCode: http.StatusUnprocessableEntity,
	}

	ErrTooManyRequests = &AppError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    "Too many requests",
		StatusCode: http.StatusTooManyRequests,
	}
)

// User errors
var (
	ErrUserNotFound = &AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "User not found",
		StatusCode: http.StatusNotFound,
	}

	ErrEmailAlreadyExists = &AppError{
		Code:       "EMAIL_EXISTS",
		Message:    "Email already registered",
		StatusCode: http.StatusConflict,
	}

	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid email or password",
		StatusCode: http.StatusUnauthorized,
	}

	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "Invalid or expired token",
		StatusCode: http.StatusUnauthorized,
	}

	ErrTokenExpired = &AppError{
		Code:       "TOKEN_EXPIRED",
		Message:    "Token has expired",
		StatusCode: http.StatusUnauthorized,
	}
)

// Account errors
var (
	ErrAccountNotFound = &AppError{
		Code:       "ACCOUNT_NOT_FOUND",
		Message:    "Account not found",
		StatusCode: http.StatusNotFound,
	}

	ErrAccountInactive = &AppError{
		Code:       "ACCOUNT_INACTIVE",
		Message:    "Account is not active",
		StatusCode: http.StatusForbidden,
	}

	ErrInsufficientBalance = &AppError{
		Code:       "INSUFFICIENT_BALANCE",
		Message:    "Insufficient balance",
		StatusCode: http.StatusBadRequest,
	}

	ErrSameAccount = &AppError{
		Code:       "SAME_ACCOUNT",
		Message:    "Cannot transfer to the same account",
		StatusCode: http.StatusBadRequest,
	}

	ErrCurrencyMismatch = &AppError{
		Code:       "CURRENCY_MISMATCH",
		Message:    "Currency mismatch between accounts",
		StatusCode: http.StatusBadRequest,
	}

	ErrInvalidAmount = &AppError{
		Code:       "INVALID_AMOUNT",
		Message:    "Invalid amount",
		StatusCode: http.StatusBadRequest,
	}
)

// Transfer errors
var (
	ErrTransferNotFound = &AppError{
		Code:       "TRANSFER_NOT_FOUND",
		Message:    "Transfer not found",
		StatusCode: http.StatusNotFound,
	}

	ErrDuplicateTransfer = &AppError{
		Code:       "DUPLICATE_TRANSFER",
		Message:    "Duplicate transfer detected",
		StatusCode: http.StatusConflict,
	}
)

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}
