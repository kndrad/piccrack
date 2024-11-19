package v1

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/kndrad/wordcrack/internal/textproc"
)

func healthCheckHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received health check request",
			slog.String("url", r.URL.String()),
		)
		w.WriteHeader(http.StatusOK)
	}
}

func handleAllWords(svc *WordService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request",
			slog.String("url", r.URL.String()),
		)
		limitParam := r.URL.Query().Get("limit")
		if limitParam == "" {
			http.Error(w, "Failed to get limit url query param", http.StatusInternalServerError)

			return
		}
		limit, err := strconv.ParseUint(limitParam, 10, 32)
		if err != nil {
			http.Error(w, "Failed to convert limit path value", http.StatusInternalServerError)

			return
		}
		if limit > math.MaxInt32 {
			http.Error(w, "Limit path value exceeds max of int32", http.StatusInternalServerError)

			return
		}
		offsetParam := r.URL.Query().Get("offset")
		if offsetParam == "" {
			http.Error(w, "Failed to get limit url query param", http.StatusInternalServerError)

			return
		}
		offset, err := strconv.ParseUint(offsetParam, 10, 32)
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
	}
}
