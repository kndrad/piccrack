package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kndrad/wcrack/config"
	"github.com/kndrad/wcrack/pkg/middleware"
)

const Version = "v1"

type server struct {
	srv *http.Server
	cfg config.HTTPConfig
	l   *slog.Logger
}

func NewServer(cfg config.HTTPConfig, svc Service, logger *slog.Logger) (*server, error) {
	if logger == nil {
		panic("logger cannot be nil")
	}

	mux := http.NewServeMux()
	const prefix = "/api/" + Version
	mux.Handle("GET "+prefix+"/healthz", healthCheckHandler(logger))
	mux.Handle("GET "+prefix+"/words", listWordsHandler(svc, logger))
	mux.Handle("POST "+prefix+"/words", createWordHandler(svc, logger))
	mux.Handle("POST "+prefix+"/words/file", uploadWordsHandler(svc, logger))
	mux.Handle("POST "+prefix+"/words/image", uploadImageWordsHandler(svc, logger))
	mux.Handle("GET "+prefix+"/words/batches", middleware.LogTime(listWordsByBatchNameHandler(svc, logger), logger))
	mux.Handle("POST "+prefix+"/phrases", middleware.LogTime(uploadImagePhrasesHandler(svc, logger), logger))

	var handler http.Handler = mux

	srv := &http.Server{
		Addr:           net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:        handler,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &server{
		srv: srv,
		cfg: cfg,
		l:   logger,
	}, nil
}

func (s *server) Start(ctx context.Context) error {
	if s == nil {
		panic("server cannot be nil")
	}

	s.l.Info("Starting to listen and serve",
		slog.String("addr", s.srv.Addr),
		slog.Bool("https_enabled", s.cfg.TLSEnabled),
	)

	errs := make(chan error, 1)

	// Start
	go func() {
		err := s.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.l.Error("Failed to listen and serve", "err", err)
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
				stop() // Stop receiving incoming signal
			case context.DeadlineExceeded:
				s.l.Info("Starting server timed out")
			}
		}
	case err := <-errs:
		if err != nil {
			return fmt.Errorf("received err: %w", err)
		}
	}

	// Shutdown
	if err := s.srv.Shutdown(ctx); err != nil {
		s.l.Error("Failed to shutdown http server", "err", err)

		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}
