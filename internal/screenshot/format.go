// Format represents the supported file formats for OCR.
// See https://tesseract-ocr.github.io/tessdoc/InputFormats.html
//
//go:generate stringer -type=Format -output=format_string.gen.go

package screenshot

type Format int

const (
	PNG Format = iota
	UNKNOWN
)

// Bytes returns the byte slice representation of the format.
func (f Format) Bytes() []byte {
	return []byte(f.String())
}
