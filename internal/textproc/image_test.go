package textproc_test

import (
	"strings"
	"testing"

	"github.com/kndrad/wordcrack/internal/textproc"
	"github.com/kndrad/wordcrack/pkg/filetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestPNGFile = "golang_0.png"
)

func TestRecognizeWords(t *testing.T) {
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

			result, err := textproc.RecognizeWords(testcase.input.content)
			require.ErrorIsf(t, err, testcase.expected.err, "expected %q but got '%q'", testcase.expected.err, err)
			require.NotNil(t, result, "expected not nil result")
		})
	}
}

func TestValidateImageSize(t *testing.T) {
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
			input:    &input{content: []byte(strings.Repeat("1", textproc.MinImageSize))},
			expected: &expected{err: nil},
		},
		"empty_err": {
			input:    &input{content: []byte{}},
			expected: &expected{err: textproc.ErrEmptyContent},
		},
		"too_small_err": {
			input:    &input{content: []byte(strings.Repeat("1", textproc.MinImageSize-1))},
			expected: &expected{err: textproc.ErrImageTooSmall},
		},
		"nil_empty_err": {
			input:    &input{content: nil},
			expected: &expected{err: textproc.ErrEmptyContent},
		},
		"too_large_err": {
			input:    &input{content: []byte(strings.Repeat("DATA", textproc.MaxImageSize+1))},
			expected: &expected{err: textproc.ErrImageTooLarge},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := textproc.ValidateImageSize(testcase.input.content)
			assert.ErrorIsf(t, err, testcase.expected.err, "expected %q but got '%q'", testcase.expected.err, err)
		})
	}
}

func TestImageFormatString(t *testing.T) {
	t.Parallel()

	type input struct {
		format textproc.Format
	}
	type expected struct {
		s string
	}
	testcases := map[string]struct {
		input    *input
		expected *expected
	}{
		"png": {
			input:    &input{format: textproc.PNG},
			expected: &expected{s: "PNG"},
		},
		"unknown": {
			input:    &input{format: textproc.UNKNOWN},
			expected: &expected{s: "UNKNOWN"},
		},
		"invalid_negative": {
			input:    &input{format: textproc.Format(-1)},
			expected: &expected{s: "Format(-1)"},
		},
		"invalid_out_of_range": {
			input: &input{format: textproc.Format(100)}, expected: &expected{s: "Format(100)"},
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

			ok := textproc.IsPNG(tc.input.content)
			assert.Equal(t, tc.expected.ok, ok)
		})
	}
}

func TestIsImageFile(t *testing.T) {
	t.Parallel()

	assert.True(t, textproc.IsImageFile("image.png"))
	assert.True(t, textproc.IsImageFile("photo.jpg"))
	assert.True(t, textproc.IsImageFile("picture.jpeg"))
	assert.False(t, textproc.IsImageFile("document.txt"))
}
