// Format represents the supported file formats for OCR.
// See https://tesseract-ocr.github.io/tessdoc/InputFormats.html
//
//go:generate stringer -type=ScreenshotFormat -output=screenshot_format_string.gen.go

package screenshot

type ScreenshotFormat int

const (
	PNG ScreenshotFormat = iota
	UNKNOWN
)

// Bytes returns the byte slice representation of the format.
func (sf ScreenshotFormat) Bytes() []byte {
	return []byte(sf.String())
}
