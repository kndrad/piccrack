package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
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
				"--path=./testdata/golang_0.png",
				"--out=" + tmpDir,
			},
		},
	}

	c := &cobra.Command{Use: "text", RunE: textCmd.RunE}
	c.Flags().AddFlagSet(textCmd.Flags())

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			buf := new(bytes.Buffer)
			c.SetOut(buf)
			c.SetErr(buf)
			c.SetArgs(tC.args)

			_, err := c.ExecuteC()
			require.NoError(t, err)

			t.Log(buf.String())
		})
	}
}
