package openf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultFileMode = 0o600
	DefaultFlags    = os.O_CREATE | os.O_RDWR
)

// Time and date format for new files.
const TimeLayout = time.RFC3339

// EntryKind represents the filesystem entry kind.
type EntryKind int

const (
	FileKind EntryKind = iota
	DirKind
	UnknownKind
)

// Checks if path is a directory or a single file with stat. Returns entry kind and an error.
func IsFileOrDir(path string) (EntryKind, error) {
	info, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return UnknownKind, fmt.Errorf("err: %w", err)
		}

		return UnknownKind, fmt.Errorf("stat %s, err: %w", path, err)
	}

	if info.IsDir() {
		return DirKind, nil
	}

	return FileKind, nil
}

func FormatTime(t time.Time, layout string) string {
	if layout == "" {
		layout = TimeLayout
	}
	return t.Format(layout)
}

type PreparedPath string

func (pp PreparedPath) String() string {
	return string(pp)
}

const DefaultExt = "txt"

// Performs path validation by checking it's kind.
// If path points to a directory - the 'default' filename (which is date string laytout
// is joined with a default extension (which is txt).
// Returns prepared path.
func PreparePath(path string, t time.Time) (PreparedPath, error) {
	path, err := RmTilde(path)
	if err != nil {
		return "", fmt.Errorf("expand path: %w", err)
	}

	entryKind, err := IsFileOrDir(path)
	if err != nil {
		return "", fmt.Errorf("file or dir err: %w", err)
	}

	switch entryKind {
	case UnknownKind:
	case DirKind:
		// When path is directory entry kind - handle new file creation.
		fname := FormatTime(t, TimeLayout)
		path = filepath.Join(path, fname+"."+DefaultExt)
	case FileKind:
	}

	return PreparedPath(path), nil
}

func Open(path string, flags int, mode os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(path, flags, mode)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	// Truncate if we're not appending
	if flags&os.O_APPEND == 0 {
		if err := f.Truncate(0); err != nil {
			return nil, fmt.Errorf("truncate file: %w", err)
		}
		if _, err := f.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("setting offset with seek: %w", err)
		}
	}

	return f, nil
}

func Join(dir, name, ext string) string {
	return filepath.Join(
		filepath.Clean(dir),
		string(filepath.Separator),
		name+"."+ext,
	)
}

func RmTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get user home dir: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	return path, nil
}
