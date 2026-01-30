package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/gobank/internal/domain/entity"
	"github.com/yourusername/gobank/internal/domain/repository"
	"github.com/yourusername/gobank/internal/infrastructure/database"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(db *database.PostgresDB) repository.UserRepository {
	return &userRepository{pool: db.Pool}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &entity.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &entity.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $2, full_name = $3, role = $4, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.FullName,
		user.Role,
	)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

type refreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(db *database.PostgresDB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{pool: db.Pool}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	return err
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()
	`
	token := &entity.RefreshToken{}
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *refreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.pool.Exec(ctx, query)
	return err
}
