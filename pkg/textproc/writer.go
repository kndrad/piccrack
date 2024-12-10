package textproc

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type Writer interface {
	Write(content []byte) (int, error)
}

func Write(w Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write words: %w", err)
	}

	return nil
}

type FileWriter struct {
	mu sync.Mutex
	f  *os.File
}

func NewFileWriter(f *os.File) *FileWriter {
	return &FileWriter{
		f: f,
	}
}

func (w *FileWriter) Write(data []byte) (int, error) {
	builder := new(strings.Builder)

	builder.Write(data)
	builder.WriteString("\n")
	n := builder.Len()

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := io.WriteString(w.f, builder.String()); err != nil {
		return 0, fmt.Errorf("failed to write: %w", err)
	}

	return n, nil
}
