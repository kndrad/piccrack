package picscan

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanImage(t *testing.T) {
	path := filepath.Join("testdata", "0.png")

	sentences, err := ScanImage(path)
	require.NoError(t, err)

	i := 0
	for range sentences {
		i++
	}

	require.Equal(t, 60, i)
}

func TestScanImages(t *testing.T) {
	path := "testdata"

	sentences, err := ScanImages(path)
	require.NoError(t, err)

	i := 0
	for range sentences {
		i++
	}

	require.Equal(t, 191, i)
}
