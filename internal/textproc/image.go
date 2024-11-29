package textproc

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/bbalet/stopwords"
	"github.com/otiai10/gosseract/v2"
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

	// ErrImageTooSmall is returned when the buffer is too small.
	ErrImageTooSmall = errors.New("buffer is too small")

	// ErrImageTooLarge is returned when the buffer exceeds the maximum allowed size.
	ErrImageTooLarge = errors.New("buffer is too large")

	// ErrUnknownImageFormat is returned when the image format is unknown.
	ErrUnknownImageFormat = errors.New("unknown image format")

	// ErrUnknownLanguage is returned when language could not be detected.
	ErrUnknownLanguage = errors.New("unknown language")
)

// OCR runs Tesseract on image content and returns cleaned from stop words
// text as a slice of bytes.
//
// Accepts options to configure the Tesseract client.
// English recognition is always applied.
//
// Content must be an image. Any other format will result in an error.
// Content size must be within allowed range. See MaxSize and MinSize.
func OCR(img []byte) ([]byte, error) {
	if err := ValidateImageSize(img); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	client := gosseract.NewClient()
	defer client.Close()

	client.Trim = true

	if err := client.SetWhitelist(Alphanumeric); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if err := client.SetImageFromBytes(img); err != nil {
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

	// Get language code to pass into stopwords.Clean
	code := lang.IsoCode639_1().String()

	words := bytes.Trim(
		stopwords.Clean([]byte(text), code, true),
		" ",
	)

	return words, nil
}

const (
	MaxImageSize int = 10 * 1024 * 1024 // 10MB
	MinImageSize     = 1 * 16           // 16B
)

// ValidateImageSize checks if the content buffer size is within the allowed size range.
func ValidateImageSize(content []byte) error {
	if len(content) == 0 {
		return errors.Wrap(ErrEmptyContent, "content is empty")
	}
	if len(content) < MinImageSize {
		return errors.Wrapf(ErrImageTooSmall, "less than %d", MinImageSize)
	}
	if len(content) > MaxImageSize {
		return errors.Wrapf(ErrImageTooLarge, "exceeds %d", MaxImageSize)
	}

	return nil
}

// IsPNG checks if the given content contains PNG metadata.
// Might be removed in the future.
func IsPNG(content []byte) bool {
	return bytes.Contains(content[:4], PNG.Bytes())
}

func IsImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}
