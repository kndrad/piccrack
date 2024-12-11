package pproc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Entry represents a file system entry with its path and content.
type Entry struct {
	path    string
	content []byte
}

// Path returns the file path of the entry.
func (e *Entry) Path() string {
	if e == nil {
		return ""
	}
	return e.path
}

// Content returns the file contents of the entry.
func (e *Entry) Content() []byte {
	if e == nil {
		return nil
	}
	return e.content
}

// FilterFunc defines a predicate for filtering file contents.
//
// Implementations should return true for contents that should be included
// in the results.
type FilterFunc func(data []byte) bool

// Filter that only checks if data is not nil.
func NoFilter(data []byte) bool {
	return data != nil
}

// Walk concurrently traverses root directory and streams filtered file entries.
// Processes regular files only.
// Caller must consume the channel to prevent leaks.
//
// Context cancellation stops the walk.
func Walk(ctx context.Context, root string, f FilterFunc) (<-chan *Entry, error) {
	c := make(chan *Entry)
	errc := make(chan error)

	go func() {
		var wg sync.WaitGroup

		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("walk: %w", err)
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				data, err := os.ReadFile(path)
				if err != nil {
					errc <- fmt.Errorf("read %s: %w", path, err)
					return
				}
				if f(data) {
					select {
					case c <- &Entry{path, data}:
					case <-ctx.Done():
						errc <- ctx.Err()
						return
					}
				}
			}()
			return nil
		})

		go func() {
			wg.Wait()
			close(c)
		}()
	}()

	if err := <-errc; err == nil {
		return c, nil
	} else {
		return nil, fmt.Errorf("err during walk: %w", err)
	}
}
