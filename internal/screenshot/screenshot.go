package screenshot

import (
	"bytes"
	"fmt"

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

	// ErrUnknownScreenshotFormat is returned when the image format is unknown.
	ErrUnknownScreenshotFormat = errors.New("unknown image format")

	// ErrUnknownLanguage is returned when language could not be detected.
	ErrUnknownLanguage = errors.New("unknown language")
)

// RecognizeText runs OCR on the provided content using Tesseract and returns cleaned from stop words
// text as a slice of bytes.
//
// Accepts options to configure the Tesseract client.
// English recognition is always applied.
//
// Content is must be a PNG image. Any other format will result in an error.
// Content size must be within allowed range. See MaxSize and MinSize.
func RecognizeText(content []byte) ([]byte, error) {
	if err := CheckSize(content); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if !IsPNG(content) {
		return nil, errors.Wrap(ErrUnknownScreenshotFormat, "decode: unknown image format")
	}

	client := tesseract.NewClient()
	defer client.Close()

	client.Trim = true

	if err := client.SetWhitelist(Alphanumeric); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if err := client.SetImageFromBytes(content); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	text, err := client.Text()
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if text == "" {
		return nil, ErrEmptyContent
	}

	languages := []lingua.Language{
		lingua.English,
		lingua.Polish,
	}

	// Build the detector with valid languages
	d := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()
	lang, exists := d.DetectLanguageOf(text)
	if !exists {
		return nil, ErrUnknownLanguage
	}
	code := lingua.GetIsoCode639_1FromValue(lang.String())

	return bytes.Trim(
		stopwords.Clean([]byte(text), code.String(), true),
		" ",
	), nil
}

// IsPNG checks if the given content contains PNG metadata.
func IsPNG(content []byte) bool {
	return bytes.Contains(content[:4], PNG.Bytes())
}

const (
	// MaxSize is the maximum allowed size for an image, set to 3 MB.
	MaxSize int = 3 * 1024 * 1024

	// MinSize is the minimum allowed size for an image, set to 1 KB.
	MinSize = 1 * 16
)

// CheckSize checks if the content buffer size is within the allowed size range.
func CheckSize(content []byte) error {
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
