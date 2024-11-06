package openf_test

import (
	"io/fs"
	"os"
	"testing"

	"github.com/kndrad/itcrack/pkg/openf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinPaths(t *testing.T) {
	t.Parallel()

	result := openf.Join("dir", "file", "json")
	expected := "dir/file.json"
	assert.Equal(t, expected, result)
}

const (
	TestTmpDir = "testtmp"
)

func TestOpenFileClean(t *testing.T) {
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

		name      string
		openFlag  int
		openPerms fs.FileMode
		mustFail  bool
	}{
		{
			desc:      "accepts_default_flag_and_perm",
			name:      tmpFile.Name(),
			openFlag:  openf.DefaultFlags,
			openPerms: openf.DefaultFileMode,
			mustFail:  false,
		},
		{
			desc:      "accepts_zero_flag",
			name:      tmpFile.Name(),
			openFlag:  0,
			openPerms: openf.DefaultFileMode,
			mustFail:  false,
		},
		{
			desc:      "accepts_zero_perm",
			name:      tmpFile.Name(),
			openFlag:  openf.DefaultFlags,
			openPerms: 0,
			mustFail:  false,
		},
		{
			desc:      "handles_directory_input",
			name:      tmpDir,
			openFlag:  openf.DefaultFlags,
			openPerms: openf.DefaultFileMode,
			mustFail:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			f, err := openf.Cleaned(tc.name, tc.openFlag, tc.openPerms)
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
