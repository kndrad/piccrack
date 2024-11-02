package cmd

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/spf13/viper"
)

const DefaultEnvFilePath = ".env"

func LoadDatabaseConfig(path string) (*screenshot.DatabaseConfig, error) {
	logger.Info("Loading database configuration")

	if path == "" {
		path = DefaultEnvFilePath
	}
	viper.SetConfigFile(filepath.Clean(path))

	if err := viper.ReadInConfig(); err != nil {
		if _, notfound := err.(viper.ConfigFileNotFoundError); notfound {
			return nil, fmt.Errorf("config file not found: %w", err)
		} else {
			return nil, fmt.Errorf("reading in config: %w", err)
		}
	}

	viper.AutomaticEnv()

	cfg := &screenshot.DatabaseConfig{
		User:     viper.GetString("DB_USER"),
		Password: viper.GetString("DB_PASSWORD"),
		Host:     viper.GetString("DB_HOST"),
		Port:     viper.GetString("DB_PORT"),
		DBName:   viper.GetString("DB_NAME"),
	}
	logger.Info("Loaded config",
		slog.String("db_host", cfg.Host),
		slog.String("db_name", cfg.DBName),
	)

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
