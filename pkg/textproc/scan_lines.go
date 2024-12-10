package textproc

import (
	"bufio"
	"strings"
	"sync"
)

func doScan(text string) <-chan string {
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

func ScanLines(texts ...string) <-chan string {
	lines := make(chan string)

	var wg sync.WaitGroup
	for _, text := range texts {
		wg.Add(1)
		go func() {
			for line := range doScan(text) {
				lines <- line
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(lines)
	}()

	return lines
}
