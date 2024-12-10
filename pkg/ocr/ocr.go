package ocr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bbalet/stopwords"
	"github.com/kndrad/wcrack/pkg/imgsniff"
	"github.com/kndrad/wcrack/pkg/pproc"
	"github.com/otiai10/gosseract/v2"
	"github.com/pemistahl/lingua-go"
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

type Word struct {
	value string
	lang  lingua.Language
}

func (w *Word) IsStop() bool {
	if w == nil {
		return true
	}
	if w.value == "" {
		return true
	}

	result := stopwords.CleanString(w.value, w.lang.IsoCode639_1().String(), false)
	return result == ""
}

func (w *Word) String() string {
	if w == nil {
		return ""
	}
	return w.value
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

func (res *Result) Words() <-chan *Word {
	var wg sync.WaitGroup

	out := make(chan *Word)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, v := range strings.Fields(res.text) {
			wg.Add(1)

			go func() {
				defer wg.Done()

				lang := detectLanguage(v)
				out <- &Word{value: strings.ToLower(v), lang: lang}
			}()
		}
	}()
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

var ErrNotAnImage = errors.New("not an image")

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

func detectLanguage(value string) lingua.Language {
	if value == "" {
		return lingua.Unknown
	}

	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(availableLanguages()...).
		Build()

	lang, exists := detector.DetectLanguageOf(value)
	if !exists {
		return 0
	}

	return lang
}

func IsImage(content []byte) bool {
	return imgsniff.IsJPG(content) || imgsniff.IsPNG(content)
}

func availableLanguages() []lingua.Language {
	return []lingua.Language{
		lingua.English,
		lingua.Polish,
	}
}

// OCR's images within a directory. Returns results.
func Dir(tc *gosseract.Client, root string) ([]*Result, error) {
	images := make([]*pproc.Entry, 0)

	entries, err := pproc.Walk(root, IsImage)
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
