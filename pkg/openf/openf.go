package openf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	path, err := expand(path)
	if err != nil {
		return nil, fmt.Errorf("expand path: %w", err)
	}

	// Check if it's a directory
	stat, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("stat: %w", err)
		}
		// If directory does not exist - create it.
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return nil, fmt.Errorf("create dir: %w", err)
		}
	} else if stat.IsDir() {
		// Make a new name for .txt file containing words if
		// the cleaned path points to a directory.
		timestamp := time.Now().Format("15_04_02_01_2006")
		dir := path
		path = filepath.Join(dir, string(filepath.Separator), timestamp+".txt")
	}

	// Continue to open
	f, err := os.OpenFile(filepath.Clean(path), flags, fm)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	// Truncate if we're not appending
	if flags&os.O_APPEND == 0 {
		if err := clearFile(f); err != nil {
			return nil, fmt.Errorf("clearing file err: %w", err)
		}
	}

	return f, nil
}

func clearFile(f *os.File) error {
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("truncate file: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("setting offset with seek: %w", err)
	}

	return nil
}

func expand(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get user home dir: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	return path, nil
}
