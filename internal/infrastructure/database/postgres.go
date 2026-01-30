package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/gobank/internal/infrastructure/config"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

func NewPostgresDB(ctx context.Context, cfg *config.DatabaseConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%d&pool_min_conns=%d",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
		cfg.MaxOpenConns,
		cfg.MaxIdleConns,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{Pool: pool}, nil
}

func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

type TxKey struct{}

func (db *PostgresDB) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	ctx = context.WithValue(ctx, TxKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(ctx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
