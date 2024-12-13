package ocr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kndrad/wcrack/pkg/imgsniff"
	"github.com/kndrad/wcrack/pkg/pproc"
	"github.com/otiai10/gosseract/v2"
)

var MaxImageSize int = 10 * 1024 * 1024 // 10MB

func NewClient() *gosseract.Client {
	client := gosseract.NewClient()
	client.Trim = true
	client.SetWhitelist(
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 \n",
	)

	return client
}

var ErrNotAnImage = errors.New("not an image")

// scan is a wrapper around tesseract client with additional content validation
// performed before returning text.
func scan(tc *gosseract.Client, content []byte) (string, error) {
	if tc == nil {
		panic("tesseract client cannot be nil")
	}

	if content == nil {
		panic("content cannot be nil")
	}
	if !IsImage(content) {
		return "", ErrNotAnImage
	}

	if err := tc.SetImageFromBytes(content); err != nil {
		return "", fmt.Errorf("set image: %w", err)
	}
	text, err := tc.Text()
	if err != nil {
		return "", fmt.Errorf("text: %w", err)
	}

	return text, nil
}

// ScanFile performs OCR on an image file.
// Image content validation is performed before ocr.
func ScanFile(tc *gosseract.Client, path string) (*Result, error) {
	if tc == nil {
		panic("tesseract client can't be nil")
	}
	if path == "" {
		panic("path can't be empty")
	}

	path = filepath.Clean(path)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	text, err := scan(tc, content)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return &Result{path: path, content: content, text: text}, nil
}

type Result struct {
	path    string
	content []byte
	text    string
}

func (res *Result) String() string {
	return res.Text()
}

func (res *Result) Text() string {
	if res == nil {
		return ""
	}

	return res.text
}

func (res *Result) Words() <-chan string {
	var wg sync.WaitGroup

	out := make(chan string)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, v := range strings.Fields(res.text) {
			wg.Add(1)

			go func() {
				defer wg.Done()

				out <- strings.ToLower(v)
			}()
		}
	}()
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// IsImage checks content (sniffs) if it's jpg or png.
func IsImage(content []byte) bool {
	return imgsniff.IsJPG(content) || imgsniff.IsPNG(content)
}

// ScanDir performs ocr on every image found in a directory.
func ScanDir(ctx context.Context, tc *gosseract.Client, root string) ([]*Result, error) {
	images := make([]*pproc.Entry, 0)

	entries, err := pproc.Walk(ctx, root, IsImage)
	if err != nil {
		return nil, fmt.Errorf("error during walk: %w", err)
	}
	for entry := range entries {
		images = append(images, entry)
	}

	// Drain entries and run ocr
	results := make([]*Result, 0)
	for _, img := range images {
		res, err := ScanFile(tc, img.Path())
		if err != nil {
			return nil, fmt.Errorf("do: %w", err)
		}
		results = append(results, res)
	}

	return results, nil
}

func ScanFrom(tc *gosseract.Client, r io.Reader) (*Result, error) {
	if tc == nil {
		return nil, errors.New("tesseract client cannot be nil")
	}
	if r == nil {
		return nil, errors.New("reader is nil")
	}

	content, err := readFull(r)
	if err != nil {
		return nil, fmt.Errorf("read full: %w", err)
	}

	text, err := scan(tc, content)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return &Result{
		path:    "",
		text:    text,
		content: content,
	}, nil
}

func readFull(r io.Reader) ([]byte, error) {
	b := make([]byte, 0, 500*1024) // Cap is 500KB (512000 bytes)

	for {
		n, err := r.Read(b[len(b):cap(b)])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("unexpected error: %w", err)
			}
		}
		if n <= 0 {
			return b, nil
		}
		b = b[:len(b)+n]
	}
}
