package pproc

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Entry struct {
	path    string
	content []byte
}

func (e *Entry) Path() string {
	if e == nil {
		return ""
	}
	return e.path
}

func (e *Entry) Content() []byte {
	if e == nil {
		return nil
	}
	return e.content
}

type FilterFunc func(data []byte) bool

// Filter that only checks if data is not nil.
func NoFilter(data []byte) bool {
	return data != nil
}

func Walk(root string, filter FilterFunc) (<-chan *Entry, error) {
	c := make(chan *Entry)
	errc := make(chan error)

	// Sends each valid image content to c
	go func() {
		var wg sync.WaitGroup

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("walk: %w", err)
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			// Checks if file content is image and sends to c
			wg.Add(1)
			go func() {
				defer wg.Done()

				data, err := os.ReadFile(path)
				if err != nil {
					errc <- fmt.Errorf("read %s: %w", path, err)
					return
				}
				if filter(data) {
					c <- &Entry{path, data}
				}
			}()
			return nil
		})

		errc <- err
		// Wait for sends to complete
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
