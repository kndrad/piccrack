package screenshot

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
)

// TextAnalysis represents a struct which contains WordFrequency field and a Name field
// of this analysis.
type TextAnalysis struct {
	name          string
	WordFrequency map[string]int `json:"wordFrequency"`

	mu sync.Mutex
}

// Creates a new TextAnalysis.
func NewTextAnalysis(name string) (*TextAnalysis, error) {
	rv, err := RandomInt(10000)
	if err != nil {
		return nil, fmt.Errorf("NewTextAnalysis: %w", err)
	}

	if name == "" {
		name = "frequency_analysis" + "_" + rv.String()
	} else {
		name = "frequency_analysis" + "_" + name + "_" + rv.String()
	}

	return &TextAnalysis{
		name:          name,
		WordFrequency: make(map[string]int),
	}, nil
}

// Adds new occurrence of a word.
// Goroutine safe.
func (ta *TextAnalysis) Add(word string) {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	ta.WordFrequency[word]++
}

func (ta *TextAnalysis) String() string {
	builder := new(strings.Builder)
	builder.WriteString(ta.name + "\n")

	for word, freq := range ta.WordFrequency {
		builder.WriteString(word + ":" + strconv.Itoa(freq) + "\n")
	}

	return builder.String()
}

func (ta *TextAnalysis) Name() string {
	return ta.name
}

var defaultX int64 = 10000

func RandomInt(x int64) (*big.Int, error) {
	if x == 0 {
		x = defaultX
	}
	i := big.NewInt(x)
	v, err := rand.Int(rand.Reader, i)
	if err != nil {
		return nil, fmt.Errorf("NewTextAnalysis: %w", err)
	}

	return v, nil
}
