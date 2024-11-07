package openf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kndrad/itcrack/internal/textproc"
)

func Join(dir, name, ext string) string {
	return filepath.Join(
		filepath.Clean(dir),
		string(filepath.Separator),
		name+"."+ext,
	)
}

const (
	DefaultFileMode = 0o600
	DefaultFlags    = os.O_CREATE | os.O_RDWR
)

func Cleaned(path string, flags int, fm fs.FileMode) (*os.File, error) {
	if flags == 0 {
		flags = DefaultFlags
	}
	if fm == 0 {
		fm = DefaultFileMode
	}

	stat, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("stat: %w", err)
		}
	} else if stat.IsDir() {
		// Make a new name for .txt file containing words if
		// the cleaned path points to a directory.
		id, err := textproc.NewAnalysisID()
		if err != nil {
			return nil, fmt.Errorf("generating analysis id: %w", err)
		}
		path = Join(path, id, "txt")
	}

	// Continue to open
	f, err := os.OpenFile(filepath.Clean(path), flags, fm)
	if err != nil {
		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("close file: %w", err)
		}

		return nil, fmt.Errorf("open file: %w", err)
	}
	if err := f.Truncate(0); err != nil {
		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("close file: %w", err)
		}

		return nil, fmt.Errorf("truncate file: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("setting offset with seek: %w", err)
	}

	return f, nil
}
