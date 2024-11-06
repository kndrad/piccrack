package cmd_test

import (
	"testing"

	"github.com/kndrad/itcrack/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnShutdown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		fn       func() error
		mustFail bool
	}{
		{
			desc: "without_err",
			fn: func() error {
				return nil
			},
			mustFail: false,
		},
		{
			desc: "returns_error",
			fn: func() error {
				return assert.AnError
			},
			mustFail: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			shutdown := cmd.OnShutdown(tc.fn)
			err := shutdown()

			switch tc.mustFail {
			case true:
				require.Error(t, err, "wanted failure (%v), but got no error: %v", tc.mustFail, err)
			case false:
				require.NoError(t, err, "wanted success (%v), but got err: %v", tc.mustFail, err)
			}
		})
	}
}
