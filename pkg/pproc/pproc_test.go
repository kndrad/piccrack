package pproc_test

import (
	"context"
	"testing"

	"github.com/kndrad/piccrack/pkg/ocr"
	"github.com/kndrad/piccrack/pkg/pproc"
	"github.com/stretchr/testify/require"
)

func TestWalkWithNoFilter(t *testing.T) {
	t.Parallel()

	root := "testdata"

	entries, err := pproc.Walk(context.Background(), root, pproc.NoFilter)
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

	entries, err := pproc.Walk(context.Background(), root, ocr.IsImage)
	require.NoError(t, err)

	for e := range entries {
		require.True(t, ocr.IsImage(e.Content()))
	}
}
