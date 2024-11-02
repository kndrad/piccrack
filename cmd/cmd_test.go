package cmd_test

import (
	"io/fs"
	"os"
	"testing"

	"github.com/kndrad/itcrack/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinPaths(t *testing.T) {
	t.Parallel()

	result := cmd.JoinPaths("dir", "file", "json")
	expected := "dir/file.json"
	assert.Equal(t, expected, result)
}

const (
	TestTmpDir = "testtmp"
)

func TestOpenCleanFile(t *testing.T) {
	t.Parallel()

	if testing.Verbose() {
		t.Logf("Creating test tmp file ")
	}

	wd, err := os.Getwd()
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp(wd, TestTmpDir)
	require.NoError(t, err)

	tmpFile, err := os.CreateTemp(tmpDir, "file*.txt")
	require.NoError(t, err)

	testCases := []struct {
		desc string

		name     string
		flag     int
		perm     fs.FileMode
		mustFail bool
	}{
		{
			desc:     "Accepts default flag and perm",
			name:     tmpFile.Name(),
			flag:     cmd.DefaultOpenFlag,
			perm:     cmd.DefaultOpenPerm,
			mustFail: false,
		},
		{
			desc:     "Accepts zero flag",
			name:     tmpFile.Name(),
			flag:     0,
			perm:     cmd.DefaultOpenPerm,
			mustFail: false,
		},
		{
			desc:     "Accepts zero perm",
			name:     tmpFile.Name(),
			flag:     cmd.DefaultOpenFlag,
			perm:     0,
			mustFail: false,
		},
		{
			desc:     "Should handle directory input",
			name:     tmpDir,
			flag:     cmd.DefaultOpenFlag,
			perm:     cmd.DefaultOpenPerm,
			mustFail: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			f, err := cmd.OpenCleanFile(tc.name, tc.flag, tc.perm)
			defer func() {
				if err := f.Close(); err != nil {
					require.NoError(t, err)
				}
			}()

			switch tc.mustFail {
			case true:
				require.Error(t, err, "test case: wanted failure (%v), but got no error: %v", tc.mustFail, err)
				assert.Nil(t, f)
			case false:
				require.NoError(t, err, "test case: wanted success (%v), but got err: %v", tc.mustFail, err)
				assert.NotNil(t, f)
			}
		})
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		require.NoError(t, err, "removing file (%s) failed: %s", tmpFile.Name(), err)
	}
	if err := os.RemoveAll(tmpFile.Name()); err != nil {
		require.NoError(t, err, "removing file (%s) failed: %s", tmpFile.Name(), err)
	}
}

func TestOnShutdown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		fn       func() error
		mustFail bool
	}{
		{
			desc: "Runs without error",
			fn: func() error {
				return nil
			},
			mustFail: false,
		},
		{
			desc: "Returns an error if a func returns an error",
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
				require.Error(t, err, "test case: wanted failure (%v), but got no error: %v", tc.mustFail, err)
			case false:
				require.NoError(t, err, "test case: wanted success (%v), but got err: %v", tc.mustFail, err)
			}
		})
	}
}
