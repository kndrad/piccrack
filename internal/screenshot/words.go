package screenshot

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type WordsFileWriter interface {
	Write(words []byte) (int, error)
}

func WriteWords(words []byte, w WordsFileWriter) error {
	// Write words
	if _, err := w.Write(words); err != nil {
		return fmt.Errorf("failed to write words: %w", err)
	}

	return nil
}

type textFileWriter struct {
	mu sync.Mutex
	f  *os.File
}

func NewTextFileWriter(f *os.File) *textFileWriter {
	return &textFileWriter{
		f: f,
	}
}

func (w *textFileWriter) Write(words []byte) (int, error) {
	builder := new(strings.Builder)

	builder.Write(words)
	builder.WriteString("\n")
	n := builder.Len()

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := io.WriteString(w.f, builder.String()); err != nil {
		return 0, fmt.Errorf("screenshot: failed to write: %w", err)
	}

	return n, nil
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
