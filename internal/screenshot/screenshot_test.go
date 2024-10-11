package screenshot_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestDataDir = "testdata"
	TestPNGFile = "golang_0.png"
	TestTmpDir  = "testtmp"
)

func Test_WriteWords(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)
	tempDirPath := filepath.Join(wd, TestDataDir)
	if !IsValidTestSubPath(t, tempDirPath) {
		t.Error("not a valid test subpath", tempDirPath)
	}

	// Create a temporary directiory for output files
	tmpDir, err := os.MkdirTemp(tempDirPath, TestTmpDir)
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a temporary tmpFile for recognition output
	tmpFile, err := os.CreateTemp(tmpDir, "out*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	words := []byte(
		"role senior golang developer crossfunctional development team engineering experiences tomorrow work",
	)

	if err := screenshot.WriteWords(words, tmpFile); err != nil {
		require.NoError(t, err)
	}
}

func Test_RecognizeContent(t *testing.T) {
	t.Parallel()

	type input struct {
		content []byte
	}
	type expected struct {
		err error
	}

	content := ReadTestPNGFile(t)

	testcases := map[string]struct {
		input    *input
		expected *expected
	}{
		"normal_input": {
			input:    &input{content: content},
			expected: &expected{err: nil},
		},
	}
	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := screenshot.RecognizeWords(testcase.input.content)
			require.ErrorIsf(t, err, testcase.expected.err, "expected %q but got '%q'", testcase.expected.err, err)
			require.NotNil(t, result, "expected not nil result")
		})
	}
}

func Test_ValidateSize(t *testing.T) {
	t.Parallel()

	type input struct {
		content []byte
	}
	type expected struct {
		err error
	}

	testcases := map[string]struct {
		input    *input
		expected *expected
	}{
		"allowed_size_ok": {
			input:    &input{content: []byte(strings.Repeat("1", screenshot.MinSize))},
			expected: &expected{err: nil},
		},
		"empty_err": {
			input:    &input{content: []byte{}},
			expected: &expected{err: screenshot.ErrEmptyContent},
		},
		"too_small_err": {
			input:    &input{content: []byte(strings.Repeat("1", screenshot.MinSize-1))},
			expected: &expected{err: screenshot.ErrTooSmall},
		},
		"nil_empty_err": {
			input:    &input{content: nil},
			expected: &expected{err: screenshot.ErrEmptyContent},
		},
		"too_large_err": {
			input:    &input{content: []byte(strings.Repeat("DATA", screenshot.MaxSize+1))},
			expected: &expected{err: screenshot.ErrTooLarge},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := screenshot.ValidateSize(testcase.input.content)
			assert.ErrorIsf(t, err, testcase.expected.err, "expected %q but got '%q'", testcase.expected.err, err)
		})
	}
}

func Test_ScreenshotFormat_String(t *testing.T) {
	t.Parallel()

	type input struct {
		format screenshot.Format
	}
	type expected struct {
		s string
	}
	testcases := map[string]struct {
		input    *input
		expected *expected
	}{
		"png": {
			input:    &input{format: screenshot.PNG},
			expected: &expected{s: "PNG"},
		},
		"unknown": {
			input:    &input{format: screenshot.UNKNOWN},
			expected: &expected{s: "UNKNOWN"},
		},
		"invalid_negative": {
			input:    &input{format: screenshot.Format(-1)},
			expected: &expected{s: "Format(-1)"},
		},
		"invalid_out_of_range": {
			input: &input{format: screenshot.Format(100)}, expected: &expected{s: "Format(100)"},
		},
	}
	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := testcase.input.format.String()
			assert.Equal(t, testcase.expected.s, s, "%s expected to be %s", s, testcase.expected.s)
		})
	}
}

func Test_IsPNG(t *testing.T) {
	t.Parallel()

	type input struct {
		content []byte
	}
	type expected struct {
		ok bool
	}

	content := ReadTestPNGFile(t)

	testcases := map[string]struct {
		input    *input
		expected *expected
	}{
		"png": {
			input:    &input{content: content},
			expected: &expected{ok: true},
		},
		"unknown": {
			input:    &input{content: []byte("very-random-metadata")},
			expected: &expected{ok: false},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ok := screenshot.IsPNG(tc.input.content)
			assert.Equal(t, tc.expected.ok, ok)
		})
	}
}

func ReadTestPNGFile(t *testing.T) []byte {
	t.Helper()

	content, err := ReadTestFile(TestPNGFile)
	if err != nil {
		t.Errorf("testIsPNG: %v", err)
	}

	return content
}

func ReadTestFile(name string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("readFile: %w", err)
	}

	name = filepath.Base(filepath.Clean(name))
	path := filepath.Join(wd, TestDataDir, name)
	if !IsSubPath(wd, path) {
		return nil, errors.New("Test file is not in the expected directory")
	}

	png, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("testFile: %w", err)
	}

	return png, nil
}

func FullTestFilePath(t *testing.T, name string) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	name = filepath.Base(filepath.Clean(name))
	path := filepath.Join(wd, TestDataDir, name)

	if !IsSubPath(wd, path) {
		t.Error("Test file is not in the expected directory")

		return ""
	}

	return path
}

func IsValidTestSubPath(t *testing.T, path string) bool {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	return IsSubPath(wd, path)
}

// IsSubPath checks if the filePath is a subpath of the base path.
func IsSubPath(basePath, filePath string) bool {
	rel, err := filepath.Rel(basePath, filePath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// func ReadImages(fsys fs.FS, dir string) [][]byte {
// 	images := make([][]byte, 0)

// 	fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
// 		ext := filepath.Ext(path)

// 		if IsSupported(ext) && !d.IsDir() {
// 			content, err := fs.ReadFile(fsys, path)
// 			if err != nil {
// 				return err
// 			}
// 			images = append(images, content)
// 		}
// 		return nil
// 	})
// 	return images
// }

// func TestDecodingManyFiles(t *testing.T) {
// 	dir, err := os.Getwd()
// 	require.Nil(t, err)

// 	fsys := os.DirFS(dir)

// 	images := ReadImages(fsys, "testdata")
// 	assert.NotEqual(t, 0, len(images))

// 	results := make([]*Result, len(images))
// 	for _, img := range images {
// 		result, err := Decode(bytes.NewBuffer(img))
// 		require.Nil(t, err)
// 		results = append(results, result)
// 	}

// 	textBuf := new(bytes.Buffer)
// 	for _, result := range results {
// 		text, err := result.Text()
// 		require.Nil(t, err)

// 		n, err := buf.Write(text)
// 		require.Nil(t, err)
// 		require.NotEqual(t, 0, n)
// 	}

// 	freq := make(map[string]int)
// 	words := strings.FieldsFunc(buf.String(), func(r rune) bool {
// 		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
// 	})
// 	for _, word := range words {
// 		freq[word]++
// 	}
// 	assert.NotEmpty(t, words)
// 	assert.NotEmpty(t, freq)
// }

// func TestWordsFrequencyAnalysis(t *testing.T) {
// 	dir, err := os.Getwd()
// 	require.Nil(t, err)
// 	path := filepath.Join(dir, "testdata", "golang_0.txt")

// 	text, err := os.ReadFile(path)
// 	require.Nil(t, err)

// 	freq := make(map[string]int)
// 	words := strings.Fields(string(text))
// 	for _, word := range words {
// 		freq[word]++
// 	}
// 	assert.NotEmpty(t, words)
// 	assert.NotEmpty(t, freq)
// }

// func TestManyFrequencyAnalysis(t *testing.T) {
// 	dir, err := os.Getwd()
// 	require.Nil(t, err)

// 	Load text from many files to buf
// 	textBuf := new(bytes.Buffer)

// 	fsys := os.DirFS(dir)
// 	t.Log("using walk dir:")
// 	fs.WalkDir(fsys, "testdata", func(path string, d fs.DirEntry, err error) error {
// 		t.Logf("path: %s, d.name: %s, err: %v", path, d.Name(), err)

// 		if strings.HasSuffix(d.Name(), ".txt") && !d.IsDir() {
// 			text, err := os.ReadFile(path)
// 			if err != nil {
// 				return err
// 			}
// 			if _, err := buf.Write(text); err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	})

// 	t.Logf("buffer: total: %d", buf.Len())

// 	freq := make(map[string]int)
// 	words := strings.FieldsFunc(buf.String(), func(r rune) bool {
// 		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
// 	})
// 	for _, word := range words {
// 		freq[word]++
// 	}
// 	assert.NotEmpty(t, words)
// 	assert.NotEmpty(t, freq)
// }
