package picphrase

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanAtPath(t *testing.T) {
	path := filepath.Join("testdata", "0.png")

	phrases, err := ScanAt(context.Background(), path)
	require.NoError(t, err)

	i := 0
	for range phrases {
		i++
	}

	require.Equal(t, 60, i)
}

func TestScanInDir(t *testing.T) {
	path := "testdata"

	phrases, err := ScanDir(context.Background(), path)
	require.NoError(t, err)

	i := 0
	for range phrases {
		i++
	}

	require.Equal(t, 191, i)
}

func TestScanReader(t *testing.T) {
	path := filepath.Join("testdata", "0.png")

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	ctx := context.Background()

	phrases, err := ScanReader(ctx, f)
	require.NoError(t, err)

	i := 0
	for range phrases {
		i++
	}

	require.Equal(t, 60, i)
}
