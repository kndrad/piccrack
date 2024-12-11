package ocr

import (
	"context"
	"errors"
	"fmt"
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

// Performs OCR on a image.
//
// Path points to an image. Image validation is performed.
//
// Returns Result.
func Do(tc *gosseract.Client, path string) (*Result, error) {
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

	if !IsImage(content) {
		return nil, ErrNotAnImage
	}

	if err := tc.SetImageFromBytes(content); err != nil {
		return nil, fmt.Errorf("set image: %w", err)
	}

	text, err := tc.Text()
	if err != nil {
		return nil, fmt.Errorf("text: %w", err)
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

// Dir is like Do but performs ocr on every image found in a directory.
//
// Returns Result slice.
func Dir(ctx context.Context, tc *gosseract.Client, root string) ([]*Result, error) {
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
		res, err := Do(tc, img.Path())
		if err != nil {
			return nil, fmt.Errorf("do: %w", err)
		}
		results = append(results, res)
	}

	return results, nil
}
