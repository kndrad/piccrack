package screenshot_test

import (
	"testing"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAnalysisName(t *testing.T) {
	t.Parallel()

	// Should return generated result.
	name, err := screenshot.GenerateAnalysisName()
	require.NoError(t, err)
	assert.NotEmpty(t, name)
}

func TestTextAnalysisName(t *testing.T) {
	t.Parallel()

	analysis := NewTestTextAnalysis(t)
	name, err := analysis.Name()
	require.NoError(t, err)
	assert.NotEmpty(t, name, "analysis name should not be empty")
}

func TestTextAnalysis_Add(t *testing.T) {
	// Test normal adding of word to the TextAnalysis WordFrequency field.
	t.Parallel()

	analysis := NewTestTextAnalysis(t)

	// Field should be initialized before
	assert.NotNil(t, analysis.WordFrequency)

	// Should add "test1" to the WordFrequency field and at first the frequency should be 1
	word := "test1"
	analysis.Add(word)
	assert.Len(t, analysis.WordFrequency, 1)
	assert.Equal(t, 1, analysis.WordFrequency[word])
	// After another "Add" it should be two
	analysis.Add(word)
	assert.Len(t, analysis.WordFrequency, 1)
	assert.Equal(t, 2, analysis.WordFrequency[word])
}

func NewTestTextAnalysis(t *testing.T) *screenshot.TextAnalysis {
	t.Helper()

	ta, err := screenshot.NewTextAnalysis()
	require.NoError(t, err)

	return ta
}
