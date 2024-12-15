package v1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kndrad/piccrack/internal/database"
	"github.com/kndrad/piccrack/pkg/ocr"
)

func limitValue(values url.Values) (int32, error) {
	var v string
	const defaultLimit = "1000"

	v = values.Get("limit")
	if v == "" {
		v = defaultLimit
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse uint: %w", err)
	}
	if n > math.MaxInt32 {
		n = 1000
	}

	return int32(n), nil
}

func offsetValue(values url.Values) (int32, error) {
	var v string
	const defaultOffset = "0"

	v = values.Get("offset")
	if v == "" {
		v = defaultOffset
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse uint: %w", err)
	}
	if n > math.MaxInt32 {
		n = 0
	}

	return int32(n), nil
}

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

func respondJSON(w http.ResponseWriter, msg string, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := struct {
		Message string `json:"message"`
		Error   string `json:"error,omitempty"`
	}{
		Message: msg,
	}
	if err != nil {
		response.Error = err.Error()
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Fprintf(w, "Internal Server Error: failed to marshal error response")
	}
}

func healthzHandler(logger *slog.Logger) http.HandlerFunc {
	type Response struct {
		Status int `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received health check request",
			slog.String("url", r.URL.String()),
		)

		resp := Response{
			Status: http.StatusOK,
		}
		if err := encode(w, r, http.StatusOK, &resp); err != nil {
			respondJSON(w, "Failed to check health", err, http.StatusInternalServerError)
		}
	}
}

func listWordsHandler(svc Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request",
			slog.String("url", r.URL.String()),
		)
		limit, err := limitValue(r.URL.Query())
		if err != nil {
			http.Error(w, "Failed get limit param from query", http.StatusBadRequest)

			return
		}
		offset, err := offsetValue(r.URL.Query())
		if err != nil {
			http.Error(w, "Failed get limit param from query", http.StatusBadRequest)

			return
		}
		rows, err := svc.ListWords(r.Context(), limit, offset)
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

func createWordHandler(svc Service, logger *slog.Logger) http.HandlerFunc {
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
		row, err := svc.CreateWord(r.Context(), request.Value)
		if err != nil {
			http.Error(w, "Failed to insert word", http.StatusInternalServerError)

			return
		}
		logger.Info("Inserted word",
			slog.Int64("id", row.ID),
			slog.String("value", row.Value),
		)
		if err := encode(w, r, http.StatusOK, row); err != nil {
			http.Error(w, "Failed to encode insert word row", http.StatusInternalServerError)

			return
		}
	}
}

func uploadWordsHandler(svc Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request",
			slog.String("url", r.URL.String()),
		)

		// Parse file from form
		const MaxSize = 1024 * 1024 * 20 // 20 MB
		r.Body = http.MaxBytesReader(w, r.Body, MaxSize)

		if err := r.ParseMultipartForm(MaxSize); err != nil {
			respondJSON(w, "File too big", err, http.StatusBadRequest)
		}
		f, fheader, err := r.FormFile("file")
		if err != nil {
			respondJSON(w, "Failed to get file", err, http.StatusBadRequest)
		}
		defer f.Close()

		// Detect content type of file and check if it's a text
		data := make([]byte, 512)
		if _, err := f.Read(data); err != nil {
			respondJSON(w, "Failed to read file into buffer", err, http.StatusInternalServerError)
		}
		contentType := http.DetectContentType(data)
		allowed := func() bool {
			types := map[string]bool{
				"text/plain":                true,
				"text/plain; charset=utf-8": true,
				"application/txt":           true,
				"text/x-plain":              true,
			}

			return types[contentType]
		}
		if !allowed() {
			respondJSON(w,
				fmt.Sprintf("Content type %s not allowed. Upload text file", contentType),
				nil,
				http.StatusBadRequest,
			)
		}
		// Return pointer back to the start of the file after content type detection
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			respondJSON(w, "Failed to seek to start of the file", err, http.StatusInternalServerError)
		}
		logger.Info("Received form",
			slog.String("filename", fheader.Filename),
		)

		// Read many words one by one from file
		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanWords)

		var words []string
		for scanner.Scan() {
			words = append(words, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			respondJSON(w, "Scanner returned an error", err, http.StatusInternalServerError)
		}
		count := 0
		for _, word := range words {
			_, err := svc.CreateWord(r.Context(), word)
			if err != nil {
				respondJSON(w, "Failed to insert row", err, http.StatusInternalServerError)
			}
			count++
		}
		type Response struct {
			Count int `json:"count"`
		}
		resp := Response{
			Count: count,
		}
		if err := encode(w, r, http.StatusOK, resp); err != nil {
			respondJSON(w, "Failed to insert row", err, http.StatusInternalServerError)
		}
		logger.Info("Inserted words", slog.Int("count", count))
	}
}

func uploadImageWordsHandler(svc Service, logger *slog.Logger) http.HandlerFunc {
	var maxSize int64 = 1024 * 1024 * 50 // 50 MB

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)

		if err := r.ParseMultipartForm(maxSize); err != nil {
			respondJSON(w, "Image file too big", err, http.StatusBadRequest)
		}
		f, header, err := r.FormFile("image")
		if err != nil {
			respondJSON(w, "Failed to get image file", err, http.StatusBadRequest)
		}
		defer f.Close()

		// Detect content type of file and check if it's a text
		data := make([]byte, 512)
		if _, err := f.Read(data); err != nil {
			respondJSON(w, "Failed to read image file into buffer", err, http.StatusInternalServerError)
		}
		contentType := http.DetectContentType(data)

		var allowed bool

		switch contentType {
		case "image/jpeg", "image/png":
			allowed = true
		default:
			allowed = false
		}
		// Return pointer back to the start of the file after content type detection
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			respondJSON(w, "Failed to seek to start of the file", err, http.StatusInternalServerError)
		}
		if !allowed {
			respondJSON(w,
				fmt.Sprintf("Content type %s not allowed. Upload text file", contentType),
				nil,
				http.StatusBadRequest,
			)
		}
		logger.Info("Received form", slog.String("header_filename", header.Filename))

		content := make([]byte, ocr.MaxImageSize)
		if _, err := f.Read(content); err != nil {
			respondJSON(w,
				"Failed to read file content for image words recognition.",
				err,
				http.StatusInternalServerError,
			)
		}
		c := ocr.NewClient()
		defer c.Close()

		result, err := ocr.ScanFile(c, header.Filename)
		if err != nil {
			respondJSON(w,
				"Failed to recognize words from an image",
				err,
				http.StatusInternalServerError,
			)
		}

		var words []string
		for w := range result.Words() {
			words = append(words, w)
		}

		row, err := svc.CreateWordsBatch(r.Context(), header.Filename, words)
		if err != nil {
			respondJSON(w, "Failed to insert words batch", err, http.StatusInternalServerError)
		}

		response := struct {
			Row database.CreateWordsBatchRow `json:"row"`
		}{
			Row: row,
		}
		if err := encode(w, r, http.StatusOK, response); err != nil {
			respondJSON(w, "Failed to encode response", err, http.StatusInternalServerError)
		}
	}
}

func listWordBatchesHandler(svc Service, logger *slog.Logger) http.HandlerFunc {
	type response struct {
		Results []database.ListWordBatchesRow `json:"word_batches"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := limitValue(r.URL.Query())
		if err != nil {
			respondJSON(w, "Failed to get limit query value", err, http.StatusBadRequest)

			return
		}
		offset, err := offsetValue(r.URL.Query())
		if err != nil {
			respondJSON(w, "Failed to get offset query value", err, http.StatusBadRequest)

			return
		}

		rows, err := svc.ListWordBatches(r.Context(), limit, offset)
		if err != nil {
			respondJSON(w, "Failed to list word batches via word service", err, http.StatusInternalServerError)

			return
		}

		resp := response{
			Results: rows,
		}
		logger.Info("Got word batches", "total", len(rows))

		if err := encode(w, r, http.StatusOK, resp); err != nil {
			respondJSON(w, "Failed to serve response", err, http.StatusInternalServerError)

			return
		}
	}
}

func listWordsByBatchNameHandler(svc Service, l *slog.Logger) http.HandlerFunc {
	type response struct {
		Rows []database.ListWordsByBatchNameRow `json:"rows"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")

		l.Info("Searching words batch", "name", name)

		rows, err := svc.ListWordsByBatchName(r.Context(), name)
		if err != nil {
			respondJSON(w, "Failed to list words by batch name with word service", err, http.StatusInternalServerError)

			return
		}
		resp := response{
			Rows: rows,
		}
		if err := encode(w, r, http.StatusOK, resp); err != nil {
			respondJSON(w, "Failed to serve response", err, http.StatusInternalServerError)

			return
		}
	}
}
