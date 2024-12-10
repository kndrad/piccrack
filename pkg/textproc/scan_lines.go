package textproc

import (
	"bufio"
	"strings"
	"sync"
)

func ScanLines(text string) <-chan string {
	out := make(chan string)

	// Scan and wait for scan to complete
	buf := bufio.NewScanner(strings.NewReader(strings.ToLower(text)))
	buf.Split(bufio.ScanLines)

	go func() {
		for buf.Scan() {
			s := strings.Trim(buf.Text(), " ")
			out <- s
		}
		close(out)
	}()

	return out
}

func ManyScanLines(texts []string) []string {
	out := make([]string, 0)

	lines := make(chan string)

	var wg sync.WaitGroup
	for _, text := range texts {
		wg.Add(1)
		go func() {
			for line := range ScanLines(text) {
				lines <- line
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(lines)
	}()

	for line := range lines {
		out = append(out, line)
	}

	return out
}
