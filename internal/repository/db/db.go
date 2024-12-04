package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/talx-hub/malerter/internal/logger/zerologger"
	"github.com/talx-hub/malerter/internal/model"
)

type DB struct {
	pool *pgxpool.Pool
	log  *zerologger.ZeroLogger
}

func New(ctx context.Context, dsn string, logger *zerologger.ZeroLogger,
) (*DB, error) {
	pool, err := initPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to init DB pool: %w", err)
	}
	return &DB{
		pool: pool,
		log:  logger,
	}, nil
}


	return nil
}

func initPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil,
			fmt.Errorf("failed to create connection pool: %w", err)
	}

	err = ping(ctx, pool)
	return pool, err
}

func (db *DB) Add(ctx context.Context, m model.Metric) error {
	return *metric, nil
}
func (db *DB) Get(ctx context.Context) ([]model.Metric, error) {
	return metrics, nil
}
func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	return ping(ctx, db.pool)
}

func ping(ctx context.Context, pool *pgxpool.Pool) error {
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping the DB: %w", err)
	}
	return nil
}
