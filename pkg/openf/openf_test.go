package openf_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kndrad/itcrack/pkg/openf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinPaths(t *testing.T) {
	t.Parallel()

	result := openf.Join("dir", "file", "json")
	expected := "dir/file.json"
	assert.Equal(t, expected, result)
}

const (
	TestTmpDir = "testtmp"
)

func TestIsFileOrDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("test"), 0o600))

	testCases := []struct {
		desc string

		path string

		mustErr  bool
		failFast bool

		expectedKind openf.EntryKind
	}{
		{
			desc: "fails_on_not_exist_and_returns_unknown_kind",

			path:         "unknown/path/idk_2039123972137219312132132121",
			expectedKind: openf.UnknownKind,

			mustErr:  true,
			failFast: true,
		},
		{
			desc: "returns_file_kind_if_path_locates_to",

			path:     tmpFile,
			mustErr:  false,
			failFast: false,

			expectedKind: openf.FileKind,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			if tC.failFast {
				_, err := openf.IsFileOrDir(tC.path)
				require.Errorf(t, err, "wanted fast error, but got: %v", err)
			}

			kind, err := openf.IsFileOrDir(tC.path)
			if tC.mustErr {
				require.Error(t, err)
				require.Equal(t, tC.expectedKind, openf.UnknownKind)
			}
			require.Exactly(t, tC.expectedKind, kind)
		})
	}
}

func TestFormatTime(t *testing.T) {
	date := time.Date(2024, 11, 9, 13, 30, 10, 0, time.UTC)
	str := openf.FormatTime(date, openf.TimeLayout)
	expected := "2024-11-09T13:30:10Z"
	require.Equal(t, expected, str)
}

func TestOpen(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	require.NoError(t, os.WriteFile(tmpFile, []byte("test"), 0o600))

	// Should return file cleaned
	f, err := openf.Open(tmpFile, openf.DefaultFlags, openf.DefaultFileMode)
	require.NoError(t, err)
	data, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	require.Empty(t, data)

	if err := f.Close(); err != nil {
		t.FailNow()
	}
}

func TestRmTilde(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		path string

		mustErr bool
	}{
		{
			desc: "returns_path_without_tidle",

			path:    "~/some/random/path",
			mustErr: false,
		},
		{
			desc: "returns_path_even_if_it_didnt_contain_tilde",

			path:    "/some/random/dir",
			mustErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			path, err := openf.RmTilde(tC.path)

			if tC.mustErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.NotContainsf(t, path, "~", "wanted no ~, but got: %s", path)
		})
	}
}

func TestPrepare(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	date := time.Date(2024, 11, 9, 13, 30, 10, 0, time.Local)

	testCases := []struct {
		desc string

		inputPath    string
		expectedPath string
	}{
		{
			desc: "should_return_path_with_a_default_file_extension_if_path_points_to_dir",

			inputPath:    tmpDir,
			expectedPath: filepath.Join(tmpDir, date.Format(time.RFC3339)+"."+openf.DefaultExt),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ppath, err := openf.PreparePath(tC.inputPath, date)
			require.NoError(t, err)
			require.Equal(t, tC.expectedPath, ppath.String())
		})
	}
}
