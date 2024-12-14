package v1

import (
	"log/slog"
	"net/http"

	"github.com/kndrad/wcrack/internal/database"
	"github.com/kndrad/wcrack/pkg/ocr"
	"github.com/kndrad/wcrack/pkg/picphrase"
)

func uploadImagePhrasesHandler(svc Service, l *slog.Logger) http.HandlerFunc {
	const maxSize int64 = 1024 * 1024 * 50 // 50 MB

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)

		if err := r.ParseMultipartForm(maxSize); err != nil {
			respondJSON(w, "File too big", err, http.StatusBadRequest)

			return
		}

		f, fh, err := r.FormFile("image")
		if err != nil {
			respondJSON(w, "Failed to get image file", err, http.StatusBadRequest)

			return
		}
		defer f.Close()

		l.Info("Received form", slog.String("header_filename", fh.Filename))

		img, err := fh.Open()
		if err != nil {
			respondJSON(w, "Failed to ocr", err, http.StatusInternalServerError)

			return
		}

		tc := ocr.NewClient()
		defer tc.Close()

		phrases, err := picphrase.ScanReader(r.Context(), img)
		if err != nil {
			respondJSON(w, "Failed to ocr", err, http.StatusInternalServerError)

			return
		}

		values := make([]string, 0)
		for phrase := range phrases {
			values = append(values, phrase.String())
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			name = fh.Filename
		}

		row, err := svc.CreatePhrasesBatch(r.Context(), name, values)
		if err != nil {
			respondJSON(w, "Failed to create phrases batch", err, http.StatusInternalServerError)

			return
		}

		response := struct {
			Name    string                         `json:"batch_name"`
			Message string                         `json:"message"`
			Row     database.CreatePhrasesBatchRow `json:"row"`
		}{
			Name:    name,
			Message: "Created phrases batch",
			Row:     row,
		}
		if err := encode(w, r, http.StatusOK, response); err != nil {
			respondJSON(w, "Failed to encode response", err, http.StatusInternalServerError)

			return
		}
	}
}
