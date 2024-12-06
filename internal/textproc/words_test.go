package textproc_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kndrad/wcrack/internal/textproc"
	"github.com/kndrad/wcrack/pkg/filetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestDataDir = "testdata"
	TestTmpDir  = "testtmp"
)

func Test_wordsTextFileWriter(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)
	tempDirPath := filepath.Join(wd, TestDataDir)
	if !IsValidTestSubPath(t, tempDirPath) {
		t.Error("not a valid test subpath", tempDirPath)
	}

	// Create a temporary directiory for output files
	tmpDir := filetest.MkTestTmpDir(t, TestDataDir, TestTmpDir)
	tmpFile := CreateTempOutTxtFile(t, tmpDir)
	defer RemoveTestFiles(tmpDir, tmpFile)

	words := []byte(
		"role senior golang developer crossfunctional development team engineering experiences tomorrow work",
	)
	if err := textproc.Write(words, textproc.NewWordsTextFileWriter(tmpFile)); err != nil {
		require.NoError(t, err)
	}

	// Assert words appeared in a tmp file
	contentInTmpFile, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, string(contentInTmpFile), string(words)+"\n")
}

func Test_WriteWords(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)
	tempDirPath := filepath.Join(wd, TestDataDir)
	if !IsValidTestSubPath(t, tempDirPath) {
		t.Error("not a valid test subpath", tempDirPath)
	}

	tmpDir := filetest.MkTestTmpDir(t, TestDataDir, TestTmpDir)
	tmpFile := CreateTempOutTxtFile(t, tmpDir)
	defer RemoveTestFiles(tmpDir, tmpFile)

	words := []byte(
		"role senior golang developer crossfunctional development team engineering experiences tomorrow work",
	)
	if err := textproc.Write(words, tmpFile); err != nil {
		require.NoError(t, err)
	}

	// Assert words appeared in a tmp file
	contentInTmpFile, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, contentInTmpFile, words)
}

// CreateTempOutTxtFile creates a temporary file for writing the outputs.
func CreateTempOutTxtFile(t *testing.T, dir string) *os.File {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	tempDirPath := filepath.Join(wd, TestDataDir)
	if !IsValidTestSubPath(t, tempDirPath) {
		t.Error("not a valid test subpath", tempDirPath)
	}

	tmpFile, err := os.CreateTemp(dir, "out*.txt")
	require.NoError(t, err)

	return tmpFile
}

func RemoveTestFiles(dir string, f *os.File) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("screenshot_test: %w", err)
	}
	if err := os.Remove(f.Name()); err != nil {
		return fmt.Errorf("screenshot_test: %w", err)
	}

	return nil
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
