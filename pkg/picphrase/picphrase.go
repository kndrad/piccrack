package picphrase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/kndrad/piccrack/pkg/ocr"
	"github.com/kndrad/piccrack/pkg/textproc"
)

type Phrase struct {
	value string
}

func (ph *Phrase) String() string {
	if ph == nil {
		return ""
	}

	return ph.value
}

// ScanAt uses ocr client to scan for phrases found in image located at path.
func ScanAt(ctx context.Context, path string) (<-chan *Phrase, error) {
	tc := ocr.NewClient()
	defer tc.Close()

	sentences := make(chan *Phrase)

	res, err := ocr.ScanFile(tc, filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("single ocr: %w", err)
	}

	var wg sync.WaitGroup
	for line := range textproc.ScanLines(res.Text()) {
		wg.Add(1)
		go func() {
			select {
			case sentences <- &Phrase{line}:
			case <-ctx.Done():
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(sentences)
	}()

	return sentences, nil
}

// ScanDir performs OCR on all images found in dir.
func ScanDir(ctx context.Context, dir string) (<-chan *Phrase, error) {
	dir = filepath.Clean(dir)

	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("path must be dir")
	}

	tc := ocr.NewClient()
	defer tc.Close()

	texts := make([]string, 0)

	results, err := ocr.ScanDir(ctx, tc, dir)
	if err != nil {
		return nil, fmt.Errorf("ocr dir: %w", err)
	}
	for _, res := range results {
		texts = append(texts, res.Text())
	}

	out := make(chan *Phrase)

	var wg sync.WaitGroup
	for _, text := range texts {
		wg.Add(1)
		go func() {
			for line := range textproc.ScanLines(text) {
				out <- &Phrase{line}
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()

	return out, nil
}

func ScanReader(ctx context.Context, r io.Reader) (<-chan *Phrase, error) {
	tc := ocr.NewClient()
	defer tc.Close()

	res, err := ocr.ScanFrom(tc, r)
	if err != nil {
		return nil, fmt.Errorf("scan from: %w", err)
	}

	out := make(chan *Phrase)

	var wg sync.WaitGroup
	for line := range textproc.ScanLines(res.Text()) {
		wg.Add(1)

		go func() {
			defer wg.Done()
			out <- &Phrase{line}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out, nil
}
