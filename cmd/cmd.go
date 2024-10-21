package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func Join(dir, filename, ext string) string {
	return filepath.Join(
		filepath.Clean(dir),
		string(filepath.Separator),
		filename+"."+ext,
	)
}

const (
	DefaultPerm = 0o600
	DefaultFlag = os.O_CREATE | os.O_RDWR
)

func OpenCleanFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	if flag == 0 {
		flag = DefaultFlag
	}
	if perm == 0 {
		perm = DefaultPerm
	}
	f, err := os.OpenFile(filepath.Clean(name), flag, perm)
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

		fmt.Println("Function terminated on shutdown completed successfully.")

		return nil
	}
}
