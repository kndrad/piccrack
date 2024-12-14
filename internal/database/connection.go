package database

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kndrad/piccrack/config"
	"github.com/pkg/errors"
)

func Pool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	hostPort := net.JoinHostPort(cfg.Host, cfg.Port)
	// example: "postgresql://username:password@localhost:5432/dbname?sslmode=disable"
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.User, cfg.Password, hostPort, cfg.Name,
	)

	pgcfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing connection string: %w", err)
	}

	// Connection pool settings
	pgcfg.MaxConns = 25
	pgcfg.MinConns = 5
	pgcfg.MaxConnLifetime = time.Hour
	pgcfg.MaxConnIdleTime = 30 * time.Minute
	// Connection timeouts
	pgcfg.ConnConfig.DialFunc = (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 5 * time.Second,
	}).DialContext

	pool, err := pgxpool.NewWithConfig(ctx, pgcfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	return pool, nil
}

func Connect(ctx context.Context, pool *pgxpool.Pool) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, pool.Config().ConnString())
	if err != nil {
		return nil, fmt.Errorf("getting db config: %w", err)
	}

	return conn, nil
}

var ErrInvalidConfig = errors.New("invalid configuration: missing required fields")

func ValidateConfig(cfg config.DatabaseConfig) error {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Name == "" {
		return ErrInvalidConfig
	}

	return nil
}
