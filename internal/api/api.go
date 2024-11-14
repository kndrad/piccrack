package api

import (
	"bytes"
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
	HTTPSEnabled bool   `mapstructure:"HTTPS_ENABLED"`
	Host         string `mapstructure:"HTTP_HOST"`
	Port         string `mapstructure:"HTTP_PORT"`
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
		HTTPSEnabled: viper.GetBool("HTTPS_ENABLED"),
		Host:         viper.GetString("HTTP_HOST"),
		Port:         viper.GetString("HTTP_PORT"),
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func (cfg *Config) Addr() string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}

type httpServer struct {
	srv    *http.Server
	cfg    *Config
	logger *slog.Logger
}

func NewHTTPServer(cfg *Config, logger *slog.Logger) *httpServer {
	if cfg == nil {
		panic("config cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:           cfg.Addr(),
		Handler:        mux,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &httpServer{
		srv:    srv,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *httpServer) Start(ctx context.Context) error {
	if s == nil {
		panic("server cannot be nil")
	}
	if s.logger == nil {
		panic("logger cannot be nil")
	}

	s.logger.Info("Starting to listen and serve",
		slog.String("addr", s.srv.Addr),
		slog.Bool("https_enabled", s.cfg.HTTPSEnabled),
	)

	errs := make(chan error, 1)

	// Start
	go func() {
		err := s.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("Failed to listen and serve", "err", err)
		}
		errs <- fmt.Errorf("listen and serve err: %w", err)
	}()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Wait
	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			switch err {
			case context.Canceled:
				s.logger.Info("Operation cancelled")
				stop() // Stop Receiving singla notifications as soon as possible
			case context.DeadlineExceeded:
				s.logger.Info("Operation timed out")
			}
		}
	case err := <-errs:
		if err != nil {
			return fmt.Errorf("received err: %w", err)
		}
	}

	// Shutdown
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown http server", "err", err)

		return fmt.Errorf("shutdown err: %w", err)
	}

	return nil
}

type httpClient struct {
	c      *http.Client
	cfg    *Config
	logger *slog.Logger
}

func NewHTTPClient(config *Config, logger *slog.Logger) *httpClient {
	return &httpClient{
		c:      &http.Client{},
		cfg:    config,
		logger: logger,
	}
}

func (c *httpClient) CheckHealth(ctx context.Context) error {
	if c == nil {
		panic("client cannot be nil")
	}

	hostPort := net.JoinHostPort(c.cfg.Host, c.cfg.Port)

	var urlPrefix string
	switch c.cfg.HTTPSEnabled {
	case true:
		urlPrefix = "https"
	case false:
		urlPrefix = "http"
	}

	url := urlPrefix + "://" + hostPort + "/" + "health"

	buffer := new(bytes.Buffer)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, buffer)
	if err != nil {
		c.logger.Error("Failed to create request", "err", err)

		return fmt.Errorf("new request err: %w", err)
	}

	c.logger.Info("Running health check", "url", url)
	resp, err := c.c.Do(req)
	if err != nil {
		c.logger.Error("Failed to do request with a client", "err", err)

		return fmt.Errorf("client do request err: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		c.logger.Info("Server OK", "statusCode", resp.StatusCode)

		return nil
	default:
		c.logger.Info("Received server response", "statusCode", resp.StatusCode)
	}

	return nil
}
