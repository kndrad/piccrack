package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kndrad/itcrack/internal/screenshot"
)

func JoinPaths(dir, name, ext string) string {
	return filepath.Join(
		filepath.Clean(dir),
		string(filepath.Separator),
		name+"."+ext,
	)
}

const (
	DefaultPerm = 0o600
	DefaultFlag = os.O_CREATE | os.O_RDWR
)

func OpenCleanFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	if flag == 0 {
		flag = DefaultFlag
	}
	if perm == 0 {
		perm = DefaultPerm
	}
	stat, err := os.Stat(path)
	if err != nil {
		logger.Error("OpenCleanFile", "err", err)

		return nil, fmt.Errorf("OpenCleanFile: %w", err)
	}

	// Make a new name for a analysis text file containing words if
	// the cleaned path points to a directory.
	if stat.IsDir() {
		filename, err := screenshot.GenerateAnalysisName()
		if err != nil {
			logger.Error("OpenCleanFile", "err", err)

			return nil, fmt.Errorf("OpenCleanFile: %w", err)
		}
		path = JoinPaths(path, filename, "txt")
	}

	f, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("OpenCleanFile: %w", err)
	}
	if err := f.Truncate(0); err != nil {
		return nil, fmt.Errorf("OpenCleanFile: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("OpenCleanFile: %w", err)
	}

	return f, nil
}

func OnShutdown(funcs ...func() error) func() error {
	return func() error {
		for _, f := range funcs {
			if err := f(); err != nil {
				return fmt.Errorf("executing func: %w", err)
			}
		}

		return nil
	}
}
