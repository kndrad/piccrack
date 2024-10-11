package screenshot

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"

	"github.com/pkg/errors"

	"github.com/bbalet/stopwords"
	tesseract "github.com/otiai10/gosseract/v2"
	"github.com/pemistahl/lingua-go"
)

const (
	// Alphanumeric is the whitelist of characters that will be recognized by Tesseract.
	Alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 \n"

	// English (English) is the default language for OCR and text language detection.
	English = "eng"
)

var (
	// ErrEmptyContent is returned when the buffer is empty.
	ErrEmptyContent = errors.New("buffer is empty")

	// ErrTooSmall is returned when the buffer is too small.
	ErrTooSmall = errors.New("buffer is too small")

	// ErrTooLarge is returned when the buffer exceeds the maximum allowed size.
	ErrTooLarge = errors.New("buffer is too large")

	// ErrUnknownFormat is returned when the image format is unknown.
	ErrUnknownFormat = errors.New("unknown image format")

	// ErrUnknownLanguage is returned when language could not be detected.
	ErrUnknownLanguage = errors.New("unknown language")
)

var logger *slog.Logger

// RecognizeWords runs OCR on the provided image content using Tesseract and returns cleaned from stop words
// text as a slice of bytes.
//
// Accepts options to configure the Tesseract client.
// English recognition is always applied.
//
// Content must be an image. Any other format will result in an error.
// Content size must be within allowed range. See MaxSize and MinSize.
func RecognizeWords(content []byte) ([]byte, error) {
	if err := ValidateSize(content); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	// if !IsPNG(content) {
	// return nil, errors.Wrap(ErrUnknownFormat, "decode: unknown image format")
	// }

	logger.Info("screenshot: launching tesseract.")
	client := tesseract.NewClient()
	defer client.Close()
	logger.Info("screenshot: tesseract client initialized.")

	client.Trim = true

	if err := client.SetWhitelist(Alphanumeric); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if err := client.SetImageFromBytes(content); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	logger.Info("screenshot: starting recoginition with tesseract client.")
	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if text == "" {
		return nil, ErrEmptyContent
	}
	logger.Info("screenshot: finisihed text recognition tesseract client.", "text_total", len(text))

	languages := []lingua.Language{
		lingua.English,
		lingua.Polish,
	}

	// Build the detector with valid languages
	logger.Info("screenshot: building language detector - detecting.")
	d := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()
	lang, exists := d.DetectLanguageOf(text)
	if !exists {
		return nil, ErrUnknownLanguage
	}
	logger.Info("screenshot: detected language of text", "lang", lang.String())

	// Get language code to pass into stopwords.Clean
	code := lang.IsoCode639_1().String()
	logger.Info("screenshot: language detector finished.", "detected_language", code)

	logger.Info("screenshot: cleaning text from stop-words.")
	words := bytes.Trim(
		stopwords.Clean([]byte(text), code, true),
		" ",
	)
	logger.Info("screenshot: finished cleaning text.", "words_total", len(words))

	return words, nil
}

const (
	// MaxSize is the maximum allowed size for an image, set to 3 MB.
	MaxSize int = 3 * 1024 * 1024

	// MinSize is the minimum allowed size for an image, set to 1 KB.
	MinSize = 1 * 16
)

// ValidateSize checks if the content buffer size is within the allowed size range.
func ValidateSize(content []byte) error {
	if len(content) == 0 {
		return errors.Wrap(ErrEmptyContent, "content is empty")
	}
	if len(content) < MinSize {
		return errors.Wrapf(ErrTooSmall, "less than %d", MinSize)
	}
	if len(content) > MaxSize {
		return errors.Wrapf(ErrTooLarge, "exceeds %d", MaxSize)
	}

	return nil
}

// IsPNG checks if the given content contains PNG metadata.
// Might be removed in the future.
func IsPNG(content []byte) bool {
	return bytes.Contains(content[:4], PNG.Bytes())
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
