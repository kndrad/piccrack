package picscan

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kndrad/wcrack/pkg/ocr"
	"github.com/kndrad/wcrack/pkg/textproc"
)

type Sentence struct {
	value string
}

func (s *Sentence) String() string {
	if s == nil {
		return ""
	}
	return s.value
}

// ScanImage uses ocr client to scan for sentences found in image located at path.
func ScanImage(ctx context.Context, path string) (<-chan *Sentence, error) {
	tc := ocr.NewClient()
	defer tc.Close()

	sentences := make(chan *Sentence)

	res, err := ocr.Do(tc, filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("single ocr: %w", err)
	}

	var wg sync.WaitGroup
	for line := range textproc.ScanLines(res.Text()) {
		wg.Add(1)
		go func() {
			select {
			case sentences <- &Sentence{line}:
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

// ScanImages performs OCR on all images found in path.
func ScanImages(ctx context.Context, path string) (<-chan *Sentence, error) {
	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("path must be dir")
	}

	tc := ocr.NewClient()
	defer tc.Close()

	texts := make([]string, 0)

	results, err := ocr.Dir(ctx, tc, path)
	if err != nil {
		return nil, fmt.Errorf("ocr dir: %w", err)
	}
	for _, res := range results {
		texts = append(texts, res.Text())
	}

	out := make(chan *Sentence)

	var wg sync.WaitGroup
	for _, text := range texts {
		wg.Add(1)
		go func() {
			for line := range textproc.ScanLines(text) {
				out <- &Sentence{line}
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
