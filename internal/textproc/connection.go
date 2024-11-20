package textproc

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	DBName   string `mapstructure:"DB_NAME"`
}

func NewDatabaseConfig(host, port, user, password, dbname string) DatabaseConfig {
	return DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
}

func LoadDatabaseConfig(path string) (*DatabaseConfig, error) {
	viper.SetConfigFile(filepath.Clean(path))

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open err: %w", err)
	}
	defer f.Close()

	if err := viper.ReadConfig(f); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := &DatabaseConfig{
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

func DatabasePool(ctx context.Context, cfg DatabaseConfig) (*pgxpool.Pool, error) {
	if err := ValidateDatabaseConfig(cfg); err != nil {
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

	return pool, nil
}

func DatabaseConnection(ctx context.Context, pool *pgxpool.Pool) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, pool.Config().ConnString())
	if err != nil {
		return nil, fmt.Errorf("getting db config: %w", err)
	}

	return conn, nil
}

var ErrInvalidConfig = errors.New("invalid configuration: missing required fields")

func ValidateDatabaseConfig(cfg DatabaseConfig) error {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.DBName == "" {
		return ErrInvalidConfig
	}

	return nil
}
