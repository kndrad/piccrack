package screenshot_test

import (
	"testing"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextAnalysis_Name(t *testing.T) {
	t.Parallel()
	// With TextAnalysis initialization,
	// name field should be initialized and Name method should return result.

	textAnalysis1 := NewTestTextAnalysis(t)
	assert.NotEmpty(t, textAnalysis1.Name())
	t.Logf("TestTextAnalysis_Name: .Name():%s", textAnalysis1.Name())
	// On empty name, the name should be still returned.
	textAnalysis2 := NewTestTextAnalysis(t)
	assert.NotEmpty(t, textAnalysis2.Name())
}

func TestTextAnalysis_Add(t *testing.T) {
	// Test normal adding of word to the TextAnalysis WordFrequency field.
	t.Parallel()

	textAnalysis := NewTestTextAnalysis(t)

	// Field should be initialized before
	assert.NotNil(t, textAnalysis.WordFrequency)

	// Should add "test1" to the WordFrequency field and at first the frequency should be 1
	word := "test1"
	textAnalysis.Add(word)
	assert.Len(t, textAnalysis.WordFrequency, 1)
	assert.Equal(t, 1, textAnalysis.WordFrequency[word])
	// After another "Add" it should be two
	textAnalysis.Add(word)
	assert.Len(t, textAnalysis.WordFrequency, 1)
	assert.Equal(t, 2, textAnalysis.WordFrequency[word])
}

func NewTestTextAnalysis(t *testing.T) *screenshot.TextAnalysis {
	t.Helper()

	ta, err := screenshot.NewTextAnalysis()
	require.NoError(t, err)

	return ta
}
