package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTextCommand(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testCases := []struct {
		desc string

		args    []string
		wantErr bool
	}{
		{
			desc: "should_analyze_the_screenshot_on_valid_file_input_and_out",

			args: []string{
				"--file=./testdata/golang_0.png",
				"--out=" + tmpDir,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			buf := new(bytes.Buffer)
			textCmd.SetOut(buf)
			textCmd.SetErr(buf)
			textCmd.SetArgs(tC.args)

			_, err := textCmd.ExecuteC()
			require.NoError(t, err)

			t.Log(buf.String())
		})
	}
}
