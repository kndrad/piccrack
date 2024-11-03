package filetest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func ReadTestPNGFile(t *testing.T, testdir, name string) []byte {
	t.Helper()

	content, err := ReadTestFile(testdir, name)
	if err != nil {
		t.Errorf("testing if file is a png: %v", err)
	}

	return content
}

func ReadTestFile(testdir, name string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting wd: %w", err)
	}

	name = filepath.Base(filepath.Clean(name))
	path := filepath.Join(wd, testdir, name)
	if !IsSub(wd, path) {
		return nil, errors.New("test file is not in the expected directory")
	}

	png, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return png, nil
}

func FullTestFilePath(t *testing.T, testdir, name string) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	name = filepath.Base(filepath.Clean(name))
	path := filepath.Join(wd, testdir, name)

	if !IsSub(wd, path) {
		t.Error("Test file is not in the expected directory")

		return ""
	}

	return path
}

func RemoveTestFiles(dir string, f *os.File) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("removing test files: %w", err)
	}
	if err := os.Remove(f.Name()); err != nil {
		return fmt.Errorf("removing test files: %w", err)
	}

	return nil
}

// CreateTempOutTxtFile creates a temporary file for writing the outputs.
func CreateTempOutTxtFile(t *testing.T, testdir, name string) *os.File {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	tempDirPath := filepath.Join(wd, testdir)
	if !IsValidTestSubPath(t, tempDirPath) {
		t.Error("not a valid test subpath", tempDirPath)
	}

	tmpFile, err := os.CreateTemp(name, "out*.txt")
	require.NoError(t, err)

	return tmpFile
}

// MkTestTmpDir creates a temporary directiory for output files.
func MkTestTmpDir(t *testing.T, testdir, name string) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	tmpDirPath := filepath.Join(wd, testdir)
	if !IsValidTestSubPath(t, tmpDirPath) {
		t.Error("not a valid test subpath", tmpDirPath)
	}

	tmpDir, err := os.MkdirTemp(tmpDirPath, name)
	require.NoError(t, err)

	return tmpDir
}

func IsValidTestSubPath(t *testing.T, path string) bool {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	return IsSub(wd, path)
}

// IsSub checks if the path is a subpath of the base path.
func IsSub(base, path string) bool {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
