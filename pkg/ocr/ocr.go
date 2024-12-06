package ocr

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/bbalet/stopwords"
	"github.com/kndrad/wcrack/pkg/imgsniff"
	"github.com/otiai10/gosseract/v2"
	"github.com/pemistahl/lingua-go"
)

var MaxImageSize int = 10 * 1024 * 1024 // 10MB

func NewClient() (*gosseract.Client, error) {
	client := gosseract.NewClient()
	client.Trim = true
	if err := client.SetWhitelist(
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 \n",
	); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return client, nil
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

type image struct {
	content []byte
}

func NewImage(content []byte) *image {
	if content == nil {
		return &image{content: []byte{}}
	}
	return &image{content: content}
}

type Result struct {
	text string
}

func (res *Result) String() string {
	if res == nil {
		return ""
	}
	return res.text
}

func (res *Result) Words() <-chan *Word {
	words := make(chan *Word)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, v := range strings.Fields(res.text) {
			wg.Add(1)

			go func() {
				defer wg.Done()

				lang := detectLanguage(v)
				words <- &Word{value: strings.ToLower(v), lang: lang}
			}()
		}
	}()
	go func() {
		wg.Wait()
		close(words)
	}()

	return words
}

var ErrNotAnImage = errors.New("not an image")

func Run(c *gosseract.Client, img *image) (*Result, error) {
	if c == nil {
		panic("tesseract client can't be nil")
	}
	if img == nil {
		panic("image can't be nil")
	}

	if !validate(img.content) {
		return nil, ErrNotAnImage
	}

	if err := c.SetImageFromBytes(img.content); err != nil {
		return nil, fmt.Errorf("set image: %w", err)
	}

	text, err := c.Text()
	if err != nil {
		return nil, fmt.Errorf("text: %w", err)
	}

	return &Result{text: text}, nil
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

func validate(content []byte) bool {
	return imgsniff.IsJPG(content) || imgsniff.IsPNG(content)
}

func availableLanguages() []lingua.Language {
	return []lingua.Language{
		lingua.English,
		lingua.Polish,
	}
}
