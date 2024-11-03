package db

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	DBName   string `mapstructure:"DB_NAME"`
}

func NewConfig(host, port, user, password, dbname string) Config {
	return Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(filepath.Clean(path))

	if err := viper.ReadInConfig(); err != nil {
		if _, notfound := err.(viper.ConfigFileNotFoundError); notfound {
			return nil, fmt.Errorf("config file not found: %w", err)
		} else {
			return nil, fmt.Errorf("reading in config: %w", err)
		}
	}

	viper.AutomaticEnv()

	cfg := &Config{
		User:     viper.GetString("DB_USER"),
		Password: viper.GetString("DB_PASSWORD"),
		Host:     viper.GetString("DB_HOST"),
		Port:     viper.GetString("DB_PORT"),
		DBName:   viper.GetString("DB_NAME"),
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func DatabasePool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	hostPort := net.JoinHostPort(cfg.Host, cfg.Port)
	// example: "postgresql://username:password@localhost:5432/dbname?sslmode=disable"
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.User, cfg.Password, hostPort, cfg.DBName,
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

	op := func() error {
		if err := ping(ctx, pool); err != nil {
			return fmt.Errorf("pinging pool: %w", err)
		}

		return nil
	}

	if err := backoff.Retry(op, backoff.NewExponentialBackOff()); err != nil {
		return nil, fmt.Errorf("retrying operation: %w", err)
	}

	return pool, nil
}

func DatabaseConnection(ctx context.Context, pool *pgxpool.Pool) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, pool.Config().ConnString())
	if err != nil {
		return nil, fmt.Errorf("getting db config: %w", err)
	}

	return conn, nil
}

func ping(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return fmt.Errorf("pinging db: %w", err)
	}

	return nil
}

var ErrInvalidConfig = errors.New("invalid configuration: missing required fields")

func validateConfig(cfg Config) error {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.DBName == "" {
		return ErrInvalidConfig
	}

	return nil
}
