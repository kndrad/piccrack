package v1

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
	Host       string `mapstructure:"HTTP_HOST"`
	Port       string `mapstructure:"HTTP_PORT"`
	TLSEnabled bool   `mapstructure:"TLS_ENABLED"`
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
		Host:       viper.GetString("HTTP_HOST"),
		Port:       viper.GetString("HTTP_PORT"),
		TLSEnabled: viper.GetBool("TLS_ENABLED"),
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func (cfg Config) Addr() string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}

func (cfg Config) BaseURL() string {
	sep := ":" + string(filepath.Separator) + string(filepath.Separator)

	var urlPrefix string
	switch cfg.TLSEnabled {
	case true:
		urlPrefix = "https"
	case false:
		urlPrefix = "http"
	}

	return urlPrefix + sep + cfg.Addr()
}

type httpServer struct {
	srv    *http.Server
	mux    *http.ServeMux
	cfg    *Config
	logger *slog.Logger
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthCheckHandler)

	return mux
}

func NewHTTPServer(config *Config, logger *slog.Logger) (*httpServer, error) {
	if config == nil {
		panic("config cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	mux := NewMux()

	srv := &http.Server{
		Addr:           config.Addr(),
		Handler:        mux,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &httpServer{
		srv:    srv,
		mux:    mux,
		cfg:    config,
		logger: logger,
	}, nil
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
		slog.Bool("https_enabled", s.cfg.TLSEnabled),
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

type Client struct {
	client *http.Client
	cfg    *Config
	logger *slog.Logger
}

func NewClient(cfg *Config, logger *slog.Logger, opts ...ClientOption) *Client {
	if cfg == nil {
		panic("config cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	c := &Client{
		client: &http.Client{},
		cfg:    cfg,
		logger: logger,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type ClientOption func(*Client)

func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.client.Timeout = d
	}
}

func (c *Client) Close() {
	c.client.CloseIdleConnections()
}

func (c *Client) Get(ctx context.Context, url string) error {
	if c == nil {
		panic("client cannot be nil")
	}
	if url == "" {
		panic("url cannot be empty")
	}

	buf := new(bytes.Buffer)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		buf,
	)
	if err != nil {
		c.logger.Error("Failed to create request", "err", err)

		return fmt.Errorf("new request err: %w", err)
	}
	c.logger.Info("Sending request",
		slog.String("url", url),
	)
	resp, err := c.client.Do(req)
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
