package screenshot

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// TextAnalysis represents a struct which contains WordFrequency field and a Name field
// of this analysis.
type TextAnalysis struct {
	name          string
	WordFrequency map[string]int `json:"wordFrequency"`

	mu sync.Mutex
}

// Creates a new TextAnalysis.
func NewTextAnalysis() (*TextAnalysis, error) {
	name, err := GenerateAnalysisName()
	if err != nil {
		return nil, fmt.Errorf("NewTextAnalysis: %w", err)
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

func (ta *TextAnalysis) Name() (string, error) {
	name := ta.name
	if name != "" {
		return name, nil
	} else {
		return GenerateAnalysisName()
	}
}

var defaultMaxInt int64 = 10000

func randomInt(x int64) (*big.Int, error) {
	if x == 0 {
		x = defaultMaxInt
	}
	i := big.NewInt(x)
	v, err := rand.Int(rand.Reader, i)
	if err != nil {
		return nil, fmt.Errorf("NewTextAnalysis: %w", err)
	}

	return v, nil
}

// Returns a string of format:
// text_analysis_randomnumber_currentdate.
func GenerateAnalysisName() (string, error) {
	// YYYY-MM-DD: 2022-03-23
	YYYYMMDD := "2006-01-02"

	date := time.Now().Format(YYYYMMDD)

	rv, err := randomInt(10000)
	if err != nil {
		return "", fmt.Errorf("NewTextAnalysisName: %w", err)
	}
	b := new(strings.Builder)
	b.WriteString("analysis")
	b.WriteString("_")
	b.WriteString(rv.String())
	b.WriteString("_")
	b.WriteString(date)

	return b.String(), nil
}
