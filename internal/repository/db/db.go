package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
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
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	pool, err := initPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to init DB pool: %w", err)
	}
	return &DB{
		pool: pool,
		log:  logger,
	}, nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
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
	const tryInsertMetricName = `INSERT INTO designation(name_designation)
VALUES ($1)
ON CONFLICT (name_designation) DO NOTHING;`

	_, err := db.pool.Exec(ctx, tryInsertMetricName, m.Name)
	if err != nil {
		return fmt.Errorf("failed to update the metric name in DB: %w", err)
	}

	const insertGauge = `INSERT INTO metric(value_metric, type_metric, name_metric)
VALUES (
    $3, 
    (SELECT id_type FROM type WHERE name_type = $2), 
    (SELECT id_designation FROM designation WHERE name_designation = $1)
)
ON CONFLICT (type_metric, name_metric) DO UPDATE
SET value_metric = EXCLUDED.value_metric;`

	const insertCounter = `INSERT INTO metric(delta_metric, type_metric, name_metric)
VALUES (
    $3, 
    (SELECT id_type FROM type WHERE name_type = $2), 
    (SELECT id_designation FROM designation WHERE name_designation = $1)
)
ON CONFLICT (type_metric, name_metric) DO UPDATE
SET value_metric = EXCLUDED.value_metric;`

	if m.Type == model.MetricTypeGauge {
		_, err = db.pool.Exec(
			ctx, insertGauge, m.Name, m.Type.String(), m.ActualValue())
	} else {
		_, err = db.pool.Exec(
			ctx, insertCounter, m.Name, m.Type.String(), m.ActualValue())
	}
	if err != nil {
		return fmt.Errorf("failed to update the metric in DB: %w", err)
	}

	return nil
}

func (db *DB) Find(ctx context.Context, typeAndName string) (model.Metric, error) {
	const findMetricQuery = `SELECT 
d.name_designation, t.name_type, m.delta_metric, m.value_metric
FROM 
	metric m
JOIN 
	designation d ON m.name_metric = d.id_designation
JOIN
	type t ON m.type_metric = t.id_type
WHERE 
	t.name_type = $1
	AND d.name_designation = $2;`

	const typePos = 0
	const namePos = 1
	result := strings.Split(typeAndName, " ")

	row := db.pool.QueryRow(
		ctx, findMetricQuery, result[typePos], result[namePos])
	metric, err := fromRow(row)
	if err != nil {
		return model.Metric{}, fmt.Errorf("failed DB query: %w", err)
	}
	return *metric, nil
}

func (db *DB) Get(ctx context.Context) ([]model.Metric, error) {
	const getAll = `SELECT 
d.name_designation, t.name_type, m.delta_metric, m.value_metric
FROM 
	metric m
JOIN 
	designation d ON m.name_metric = d.id_designation
JOIN
	type t ON m.type_metric = t.id_type`

	rows, err := db.pool.Query(ctx, getAll)
	if err != nil {
		return nil, fmt.Errorf("failed to query DB: %w", err)
	}
	defer rows.Close()

	metrics := make([]model.Metric, 0)
	for rows.Next() {
		m, err := fromRow(rows)
		if err != nil {
			db.log.Error().
				Err(err).
				Msg("row error")
			continue
		}
		metrics = append(metrics, *m)
	}

	return metrics, nil
}

func fromRow(row pgx.Row) (*model.Metric, error) {
	var t string
	var metric model.Metric
	if err := row.Scan(
		&metric.Name,
		&t,
		&metric.Delta,
		&metric.Value,
	); err != nil {
		return nil, fmt.Errorf("failed to scan a response row: %w", err)
	}
	metric.Type = model.MetricType(t)

	return &metric, nil
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
