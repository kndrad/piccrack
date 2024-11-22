package v1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

func encode[T any](w http.ResponseWriter, _ *http.Request, status int, v T) error {
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
		offset, err := getOffset(r.URL.Query())
		if err != nil {
			http.Error(w, "Failed get limit param from query", http.StatusBadRequest)

			return
		}
		rows, err := svc.GetAllWords(r.Context(), limit, offset)
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

// TODO: TEST THIS FUNCTION with curl!
func insertWordsFromFileHandler(svc *WordService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request",
			slog.String("url", r.URL.String()),
		)
		// Open file
		// But should it get path from query?
		path := r.URL.Query().Get("file")
		if path == "" {
			writeJsonErr(w, "Failed to get url file path from query", nil, http.StatusBadRequest)

			return
		}
		txtf, err := os.Open(filepath.Clean(path))
		if err != nil {
			writeJsonErr(w, "Failed to open file", err, http.StatusInternalServerError)

			return
		}
		defer txtf.Close()

		// Read words one by one from txt file
		scanner := bufio.NewScanner(txtf)
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			// Insert
			row, err := svc.InsertWord(r.Context(), scanner.Text())
			if err != nil {
				writeJsonErr(w, "Failed to insert row", err, http.StatusInternalServerError)

				return
			}
			logger.Info("Inserted word",
				slog.Int64("id", row.ID),
				slog.String("value", row.Value),
			)
		}
		if err := scanner.Err(); err != nil {
			writeJsonErr(w, "Scanner returned an error", err, http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

// TODO: TEST THIS FUNCTION
func writeJsonErr(w http.ResponseWriter, msg string, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(w, fmt.Sprintf("%s | err: %v", msg, err), code)
	} else {
		http.Error(w, msg, code)
	}
}
