package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Host string `mapstructure:"HTTP_HOST"`
	Port string `mapstructure:"HTTP_PORT"`
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

	config := &Config{
		Host: viper.GetString("HTTP_HOST"),
		Port: viper.GetString("HTTP_PORT"),
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return config, nil
}

func (cfg *Config) Addr() string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}

func NewHTTPServer(c *Config, l *slog.Logger) *http.Server {
	if l == nil {
		panic("logger cannot be nil")
	}
	if c == nil {
		panic("config cannot be nil")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return &http.Server{
		Addr:           c.Addr(),
		Handler:        mux,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func StartServer(ctx context.Context, srv *http.Server, logger *slog.Logger) error {
	if srv == nil {
		panic("server cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	logger.Info("Starting to listen and serve", "addr", srv.Addr)

	errs := make(chan error, 1)

	// Start
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to listen and serve", "err", err.Error())
		}
		errs <- err
	}()

	// Wait
	select {
	case <-ctx.Done():
		logger.Info("Cancelled")
	case sig := <-signals:
		logger.Info("Received signal", slog.String("sig_string", sig.String()))
	case err := <-errs:
		if err != nil {
			return fmt.Errorf("received err: %w", err)
		}
	}

	// Shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown http server", "err", err)

		return fmt.Errorf("shutdown err: %w", err)
	}

	return nil
}
