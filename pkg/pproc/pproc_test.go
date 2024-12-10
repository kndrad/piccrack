package pproc_test

import (
	"testing"

	"github.com/kndrad/wcrack/pkg/ocr"
	"github.com/kndrad/wcrack/pkg/pproc"
	"github.com/stretchr/testify/require"
)

func TestWalkWithNoFilter(t *testing.T) {
	t.Parallel()

	root := "testdata"

	entries, err := pproc.Walk(root, pproc.NoFilter)
	require.NoError(t, err)

	total := 0
	for range entries {
		total++
	}
	require.Equal(t, 5, total)
}

func TestWalkWithImageFilter(t *testing.T) {
	t.Parallel()

	root := "testdata"

	entries, err := pproc.Walk(root, ocr.IsImage)
	require.NoError(t, err)

	for e := range entries {
		require.True(t, ocr.IsImage(e.Content()))
	}
}
