package picscan

import (
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

func ScanImage(path string) (<-chan *Sentence, error) {
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
			sentences <- &Sentence{line}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(sentences)
	}()

	return sentences, nil
}

func ScanImages(path string) (<-chan *Sentence, error) {
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

	results, err := ocr.Dir(tc, path)
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
