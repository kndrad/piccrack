package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/kndrad/itcrack/internal/textproc"
)

func healthCheckHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received health check request",
			slog.String("url", r.URL.String()),
		)
		w.WriteHeader(http.StatusOK)
	}
}

type WordsService struct {
	q      textproc.Querier
	logger *slog.Logger
}

func NewWordsService(queries textproc.Querier, logger *slog.Logger) *WordsService {
	return &WordsService{
		q:      queries,
		logger: logger,
	}
}

func (svc *WordsService) AllWords(ctx context.Context, limit, offset int32) ([]textproc.AllWordsRow, error) {
	rows, err := svc.q.AllWords(ctx, textproc.AllWordsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("query all words, err: %w", err)
	}

	return rows, nil
}

func handleAllWords(svc *WordsService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request",
			slog.String("url", r.URL.String()),
		)
		logger.Info("Connecting to database")

		limit, err := strconv.ParseUint(r.PathValue("limit"), 10, 32)
		if err != nil {
			http.Error(w, "Failed to convert limit path value", http.StatusInternalServerError)

			return
		}
		if limit > math.MaxInt32 {
			http.Error(w, "Limit path value exceeds max of int32", http.StatusInternalServerError)

			return
		}
		offset, err := strconv.ParseUint(r.PathValue("offset"), 10, 32)
		if err != nil {
			http.Error(w, "Failed to convert offset path value", http.StatusInternalServerError)

			return
		}
		if offset > math.MaxInt32 {
			http.Error(w, "Offset path value exceeds max of int32", http.StatusInternalServerError)

			return
		}
		rows, err := svc.q.AllWords(r.Context(), textproc.AllWordsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			http.Error(w, "Failed to fetch all words from a database", http.StatusInternalServerError)

			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(rows); err != nil {
			http.Error(w, "Failed to encode rows", http.StatusInternalServerError)

			return
		}

		// FIXME:
		// http: superfluous response.WriteHeader call from v1.NewServer.handleAllWords.func2 (handlers.go:92)
		w.WriteHeader(http.StatusOK)
	}
}
