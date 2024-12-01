package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const Version = "v1"

type server struct {
	srv *http.Server
	cfg Config
	l   *slog.Logger
}

func NewServer(cfg Config, wordService WordService, logger *slog.Logger) (*server, error) {
	if logger == nil {
		panic("logger cannot be nil")
	}

	mux := http.NewServeMux()
	const prefix = "/api/" + Version
	mux.Handle("GET "+prefix+"/healthz", healthCheckHandler(logger))
	mux.Handle("GET "+prefix+"/words", listWordsHandler(wordService, logger))
	mux.Handle("POST "+prefix+"/words", createWordHandler(wordService, logger))
	mux.Handle("POST "+prefix+"/words/file", uploadWordsHandler(wordService, logger))
	mux.Handle("POST "+prefix+"/words/image", uploadImageWordsHandler(wordService, logger))

	var handler http.Handler = mux

	srv := &http.Server{
		Addr:           cfg.Addr(),
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

func (srv *server) Start(ctx context.Context) error {
	if srv == nil {
		panic("server cannot be nil")
	}

	srv.l.Info("Starting to listen and serve",
		slog.String("addr", srv.srv.Addr),
		slog.Bool("https_enabled", srv.cfg.TLSEnabled),
	)

	errs := make(chan error, 1)

	// Start
	go func() {
		err := srv.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			srv.l.Error("Failed to listen and serve", "err", err)
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
				srv.l.Info("Operation cancelled")
				stop() // Stop Receiving singla notifications as soon as possible
			case context.DeadlineExceeded:
				srv.l.Info("Operation timed out")
			}
		}
	case err := <-errs:
		if err != nil {
			return fmt.Errorf("received err: %w", err)
		}
	}

	// Shutdown
	if err := srv.srv.Shutdown(ctx); err != nil {
		srv.l.Error("Failed to shutdown http server", "err", err)

		return fmt.Errorf("shutdown err: %w", err)
	}

	return nil
}
