package imgsniff

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPNG(t *testing.T) {
	t.Parallel()

	pngSignature := []byte{137, 80, 78, 71, 13, 10, 26, 10}

	testCases := []struct {
		desc string

		data []byte
		want bool
	}{
		{
			desc: "without_whitespace",

			data: []byte{137, 80, 78, 71, 13, 10, 26, 10},
			want: true,
		},
		{
			desc: "also_is_png_even_when_begins_with_whitespaces",

			data: slices.Concat([]byte(" "), pngSignature),
			want: true,
		},
		{
			desc: "is_png_even_when_with_many_whitespaces",

			data: slices.Concat([]byte("      "), pngSignature),
			want: true,
		},
		{
			desc: "is_not_png",

			data: []byte{138, 10, 20, 30, 30, 20, 32, 15}, // Some random numbers
			want: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			ok := IsPNG(tC.data)
			require.Equal(t, tC.want, ok)
		})
	}
}

func TestFirstNonWSIdx(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		data []byte
		want int
	}{
		{
			desc: "no_whitespace_bytes",

			data: []byte{137, 80, 78, 71, 13, 10, 26, 10},
			want: 0,
		},

		{
			desc: "with_whitespace",

			data: slices.Concat([]byte(" "), []byte{137, 80, 78}),
			want: 1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			idx := firstNonWSIndex(tC.data)
			require.Equal(t, tC.want, idx)
		})
	}
}

func TestIsJPG(t *testing.T) {
	t.Parallel()

	signature := []byte{0xFF, 0x4F, 0xFF, 0x51}

	testCases := []struct {
		desc string

		data []byte
		want bool
	}{
		{
			desc: "normal_jpg",

			data: []byte{0xFF, 0x4F, 0xFF, 0x51},
			want: true,
		},
		{
			desc: "first_is_whitespace_still_jpg",

			data: slices.Concat([]byte(" "), signature),
			want: true,
		},
		{
			desc: "many_whitespaces_still_jpg",

			data: slices.Concat([]byte("      "), signature),
			want: true,
		},
		{
			desc: "is_not_jpg",

			data: []byte{138, 10, 20, 30}, // Some random numbers
			want: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			t.Parallel()

			ok := IsJPG(tC.data)
			require.Equal(t, tC.want, ok)
		})
	}
}
