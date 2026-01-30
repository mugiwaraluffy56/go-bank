package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserInput struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
}

type UpdateUserInput struct {
	FullName string `json:"full_name" validate:"omitempty,min=2,max=255"`
	Email    string `json:"email" validate:"omitempty,email,max=255"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RefreshToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUser(email, passwordHash, fullName string) *User {
	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
