package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	HTTP     API            `mapstructure:"http"`
	App      AppConfig      `mapstructure:"app"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")
	v.SetConfigFile(filepath.Clean(path))

	v.SetDefault("Database.Pool.MaxConns", 25)
	v.SetDefault("Database.Pool.MinConns", 5)
	v.SetDefault("Database.Pool.MaxConnLifetime", "1h")
	v.SetDefault("Database.Pool.MaxConnIdleTime", "30m")
	v.SetDefault("Database.Pool.ConnectTimeout", "10s")
	v.SetDefault("Database.Pool.DialerKeepAlive", "5s")

	v.SetDefault("HTTP.Host", "0.0.0.0")
	v.SetDefault("HTTP.Port", "8080")

	v.SetDefault("App.Environment", "development")
	v.SetDefault("App.LogLevel", "info")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

type AppConfig struct {
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	User     string     `mapstructure:"user"`
	Password string     `mapstructure:"password"`
	Host     string     `mapstructure:"host"`
	Port     string     `mapstructure:"port"`
	Name     string     `mapstructure:"name"`
	Pool     poolConfig `mapstructure:"pool"`
}

type poolConfig struct {
	MaxConns        int    `mapstructure:"max_conns"`
	MinConns        int    `mapstructure:"min_conns"`
	MaxConnLifetime string `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime string `mapstructure:"max_conn_idle_time"`
	ConnectTimeout  string `mapstructure:"connect_timeout"`
	DialerKeepAlive string `mapstructure:"dialer_keep_alive"`
}

type API struct {
	Host       string `mapstructure:"host"`
	Port       string `mapstructure:"port"`
	TLSEnabled bool   `mapstructure:"tls_enabled"`
}
