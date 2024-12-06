package imgsniff

import (
	"bytes"
	"sync"
	"unicode"
)

type signature []byte

type signatures struct {
	m  map[string]signature
	mu sync.Mutex
}

func Signatures() *signatures {
	return &signatures{
		m: map[string]signature{
			"png": {137, 80, 78, 71, 13, 10, 26, 10},
			"jpg": {0xFF, 0x4F, 0xFF, 0x51},
		},
	}
}

func (s *signatures) png() signature {
	if s == nil {
		s = Signatures()
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.m["png"]
}

func IsPNG(data []byte) bool {
	if data == nil {
		return false
	}

	idx := firstNonWSIndex(data)
	sigs := Signatures()

	return bytes.Equal(data[idx:idx+8], sigs.png())
}

func firstNonWSIndex(data []byte) int {
	if data == nil {
		return 0
	}
	idx := 0 // Index of first non-whitespace byte in data
	for ; idx < len(data) && unicode.IsSpace(rune(data[idx])); idx++ {
	}

	return idx
}

func (s *signatures) jpg() signature {
	if s == nil {
		s = Signatures()
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.m["jpg"]
}

func IsJPG(data []byte) bool {
	if data == nil {
		return false
	}

	idx := firstNonWSIndex(data)
	sigs := Signatures()

	return bytes.Equal(data[idx:idx+4], sigs.jpg())
}
