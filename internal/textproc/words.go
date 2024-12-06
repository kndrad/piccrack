package textproc

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type WordsWriter interface {
	Write(words []byte) (int, error)
}

func Write(data []byte, w WordsWriter) error {
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write words: %w", err)
	}

	return nil
}

type wordsTextFileWriter struct {
	mu sync.Mutex
	f  *os.File
}

func NewWordsTextFileWriter(f *os.File) *wordsTextFileWriter {
	return &wordsTextFileWriter{
		f: f,
	}
}

func (w *wordsTextFileWriter) Write(words []byte) (int, error) {
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
