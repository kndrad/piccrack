package screenshot_test

import (
	"strings"
	"testing"

	"github.com/kndrad/wordcrack/internal/screenshot"
	"github.com/kndrad/wordcrack/pkg/filetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestPNGFile = "golang_0.png"
	TestDataDir = "testdata"
)

func Test_RecognizeContent(t *testing.T) {
	t.Parallel()

	type input struct {
		content []byte
	}
	type expected struct {
		err error
	}

	content := filetest.ReadTestPNGFile(t, TestDataDir, TestPNGFile)

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

	content := filetest.ReadTestPNGFile(t, TestDataDir, TestPNGFile)

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

func TestIsImageFile(t *testing.T) {
	t.Parallel()

	assert.True(t, screenshot.IsImageFile("image.png"))
	assert.True(t, screenshot.IsImageFile("photo.jpg"))
	assert.True(t, screenshot.IsImageFile("picture.jpeg"))
	assert.False(t, screenshot.IsImageFile("document.txt"))
}
