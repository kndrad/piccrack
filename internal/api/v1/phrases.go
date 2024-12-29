package v1

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kndrad/piccrack/pkg/picphrase"
)

func uploadImagePhrasesHandler(svc Service, l *slog.Logger) http.HandlerFunc {
	const maxSize int64 = 1024 * 1024 * 50 // 50 MB

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)

		if err := r.ParseMultipartForm(maxSize); err != nil {
			respondJSON(w, "File too big", err, http.StatusBadRequest)

			return
		}

		file, header, err := r.FormFile("image")
		if err != nil {
			respondJSON(w, "Failed to get image file", err, http.StatusBadRequest)

			return
		}
		defer file.Close()

		l.Info("Received form", slog.String("header_filename", header.Filename))

		img, err := header.Open()
		if err != nil {
			respondJSON(w, "Failed to ocr", err, http.StatusInternalServerError)

			return
		}

		phrases, err := picphrase.ScanReader(r.Context(), img)
		if err != nil {
			respondJSON(w, "Failed to ocr", err, http.StatusInternalServerError)

			return
		}

		values := make([]string, 0)
		for phrase := range phrases {
			values = append(values, phrase.String())
		}

		name := strings.Split(header.Filename, ".")[0] + "_" + time.Now().Format("20060102_150405")
		row, err := svc.CreatePhrasesBatch(r.Context(), name, values)
		if err != nil {
			respondJSON(w, "Failed to create phrases batch", err, http.StatusInternalServerError)

			return
		}

		response := struct {
			Message string `json:"message"`
		}{
			Message: fmt.Sprintf(
				"Created phrases batch with a name: %s and id: %d\n", name, row.ID),
		}
		if err := encode(w, r, http.StatusOK, response); err != nil {
			respondJSON(w, "Failed to encode response", err, http.StatusInternalServerError)

			return
		}
	}
}
