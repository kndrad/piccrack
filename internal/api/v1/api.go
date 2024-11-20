package v1

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

const Version = "v1"

type ServerConfig struct {
	Host       string `mapstructure:"HTTP_HOST"`
	Port       string `mapstructure:"HTTP_PORT"`
	TLSEnabled bool   `mapstructure:"TLS_ENABLED"`
}

func LoadConfig(path string) (*ServerConfig, error) {
	viper.SetConfigFile(filepath.Clean(path))

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open err: %w", err)
	}
	defer f.Close()

	if err := viper.ReadConfig(f); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := &ServerConfig{
		Host:       viper.GetString("HTTP_HOST"),
		Port:       viper.GetString("HTTP_PORT"),
		TLSEnabled: viper.GetBool("TLS_ENABLED"),
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func (cfg ServerConfig) Addr() string {
	return net.JoinHostPort(cfg.Host, cfg.Port)
}

func (cfg ServerConfig) BaseURL() string {
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

type Server struct {
	config *ServerConfig
	logger *slog.Logger

	srv *http.Server
}

func NewServer(config *ServerConfig, wordsService *WordService, logger *slog.Logger) (*Server, error) {
	if config == nil {
		panic("config cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}

	mux := http.NewServeMux()
	const prefix = "/api/" + Version
	mux.Handle("GET "+prefix+"/healthz", healthCheckHandler(logger))
	mux.Handle("GET "+prefix+"/words", allWordsHandler(wordsService, logger))
	mux.Handle("POST "+prefix+"/words", insertWordHandler(wordsService, logger))

	var handler http.Handler = mux

	srv := &http.Server{
		Addr:           config.Addr(),
		Handler:        handler,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &Server{
		srv:    srv,
		config: config,
		logger: logger,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	if s == nil {
		panic("server cannot be nil")
	}
	if s.logger == nil {
		panic("logger cannot be nil")
	}

	s.logger.Info("Starting to listen and serve",
		slog.String("addr", s.srv.Addr),
		slog.Bool("https_enabled", s.config.TLSEnabled),
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
