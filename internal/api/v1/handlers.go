package v1

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	return v, nil
}

func healthCheckHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received health check request",
			slog.String("url", r.URL.String()),
		)
		w.WriteHeader(http.StatusOK)
	}
}

func getLimit(values url.Values) (int32, error) {
	var param string
	const defaultLimitParam = "1000"

	param = values.Get("limit")
	if param == "" {
		param = defaultLimitParam
	}
	limit, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse uint: %w", err)
	}
	if limit > math.MaxInt32 {
		limit = 1000
	}

	return int32(limit), nil
}

func getOffset(values url.Values) (int32, error) {
	var param string
	const defaultOffsetParam = "0"

	param = values.Get("offset")
	if param == "" {
		param = defaultOffsetParam
	}
	offset, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse uint: %w", err)
	}
	if offset > math.MaxInt32 {
		offset = 0
	}

	return int32(offset), nil
}

func allWordsHandler(svc *WordService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request",
			slog.String("url", r.URL.String()),
		)
		limit, err := getLimit(r.URL.Query())
		if err != nil {
			http.Error(w, "Failed get limit param from query", http.StatusBadRequest)

			return
		}
		offsetParam := r.URL.Query().Get("offset")
		if offsetParam == "" {
			http.Error(w, "Failed to get limit url query param", http.StatusBadRequest)

			return
		}
		offset, err := strconv.ParseUint(offsetParam, 10, 32)
		if err != nil {
			http.Error(w, "Failed to convert offset path value", http.StatusBadRequest)

			return
		}
		if offset > math.MaxInt32 {
			http.Error(w, "Offset path value exceeds max of int32", http.StatusBadRequest)

			return
		}
		rows, err := svc.GetAllWords(r.Context(), limit, int32(offset))
		if err != nil {
			http.Error(w, "Failed to fetch all words from a database", http.StatusInternalServerError)

			return
		}
		if err := encode(w, r, http.StatusOK, rows); err != nil {
			http.Error(w, "Failed to encode rows", http.StatusInternalServerError)

			return
		}
	}
}

func insertWordHandler(svc *WordService, logger *slog.Logger) http.HandlerFunc {
	type Request struct {
		Value string `json:"value"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request",
			slog.String("url", r.URL.String()),
		)

		request, err := decode[Request](r)
		if err != nil {
			http.Error(w, "Failed to decode request", http.StatusBadRequest)

			return
		}
		row, err := svc.InsertWord(r.Context(), request.Value)
		if err != nil {
			http.Error(w, "Failed to insert word", http.StatusInternalServerError)

			return
		}
		logger.Info("Word inserted",
			slog.Int64("id", row.ID),
			slog.String("value", row.Value),
		)
		if err := encode(w, r, http.StatusOK, row); err != nil {
			http.Error(w, "Failed to encode insert word row", http.StatusInternalServerError)

			return
		}
	}
}
