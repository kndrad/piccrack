package screenshot_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kndrad/reckon/internal/screenshot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestDataDir = "testdata"
	TestPNGFile = "golang_0.png"
)

func TestRecognizeText(t *testing.T) {
	t.Parallel()

	type input struct {
		content []byte
	}
	type expected struct {
		err error
	}

	path := filepath.Join(TestDataDir, TestPNGFile)
	content, err := os.ReadFile(path)
	require.NoError(t, err)

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

			result, err := screenshot.RecognizeText(testcase.input.content)
			require.ErrorIsf(t, err, testcase.expected.err, "expected %q but got '%q'", testcase.expected.err, err)
			require.NotNil(t, result, "expected not nil result")
		})
	}
}

func TestValidateBufferSize(t *testing.T) {
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

			err := screenshot.CheckSize(testcase.input.content)
			assert.ErrorIsf(t, err, testcase.expected.err, "expected %q but got '%q'", testcase.expected.err, err)
		})
	}
}

func TestScreenshotFormat_String(t *testing.T) {
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
			expected: &expected{s: "ScreenshotFormat(-1)"},
		},
		"invalid_out_of_range": {
			input: &input{format: screenshot.Format(100)}, expected: &expected{s: "ScreenshotFormat(100)"},
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

func TestIsPNG(t *testing.T) {
	t.Parallel()

	type input struct {
		content []byte
	}
	type expected struct {
		ok bool
	}

	wd, err := os.Getwd()
	require.NoError(t, err)

	path := filepath.Join(wd, TestDataDir, TestPNGFile)
	if !isSubPath(wd, path) {
		t.Fatalf("Test file is not in the expected directory")
	}

	png, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)

	testcases := map[string]struct {
		input    *input
		expected *expected
	}{
		"png": {
			input:    &input{content: png},
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

// isSubPath checks if the filePath is a subpath of the base path.
func isSubPath(basePath, filePath string) bool {
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

// 	buf := bytes.Buffer{}
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
// 	buf := bytes.Buffer{}

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
