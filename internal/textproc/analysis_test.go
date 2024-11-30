package textproc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kndrad/wcrack/internal/textproc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAnalysisID(t *testing.T) {
	t.Parallel()

	// Should return generated result.
	name, err := textproc.NewAnalysisID()
	require.NoError(t, err)
	assert.NotEmpty(t, name)
}

func TestTextAnalysisID(t *testing.T) {
	t.Parallel()

	analysis := NewTestTextAnalysis(t)
	id := analysis.ID
	assert.NotEmpty(t, id)
}

func TestTextAnalysis_Add(t *testing.T) {
	// Test normal adding of word to the TextAnalysis WordFrequency field.
	t.Parallel()

	analysis := NewTestTextAnalysis(t)

	// Field should be initialized before
	assert.NotNil(t, analysis.WordFrequency)

	// Should add "test1" to the WordFrequency field and at first the frequency should be 1
	word := "test1"
	analysis.IncWordCount(word)
	assert.Len(t, analysis.WordFrequency, 1)
	assert.Equal(t, 1, analysis.WordFrequency[word])
	// After another "Add" it should be two
	analysis.IncWordCount(word)
	assert.Len(t, analysis.WordFrequency, 1)
	assert.Equal(t, 2, analysis.WordFrequency[word])
}

func NewTestTextAnalysis(t *testing.T) *textproc.TextAnalysis {
	t.Helper()

	ta, err := textproc.NewTextAnalysis()
	require.NoError(t, err)

	return ta
}

func TestWordsFrequencyAnalysis(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		words    []string
		mustFail bool
	}{
		{
			desc:     "Nil words input fails",
			words:    nil,
			mustFail: true,
		},
		{
			desc:     "Success on analysing words read from a file",
			words:    NewTestWords(t),
			mustFail: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			analysis, err := textproc.AnalyzeFrequency(tc.words)

			if tc.mustFail {
				require.Error(t, err, "wanted failure but got: %w", err)
			} else {
				require.NoError(t, err, "wanted success but got: %w", err)
				assert.NotNil(t, analysis)
			}
		})
	}
}

func NewTestWords(t *testing.T) []string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("NewTestWords: %v", err)

		return nil
	}
	sep := string(filepath.Separator)
	name := "words.txt"
	path := filepath.Join(
		wd+sep,
		"testdata"+sep,
		name,
	)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("NewTestWords: %v", err)

		return nil
	}
	buffer := bytes.NewBuffer(data)

	return strings.Split(buffer.String(), " ")
}

func TestGeneratingAnalysisIDWithSuffix(t *testing.T) {
	t.Parallel()

	id, err := textproc.NewAnalysisIDWithSuffix("dir")
	require.NoError(t, err)
	assert.Contains(t, id, "dir_")
}
