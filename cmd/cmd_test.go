package cmd_test

import (
	"testing"

	"github.com/kndrad/itcrack/cmd"
	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	t.Parallel()

	result := cmd.Join("dir", "file", "json")
	expected := "dir/file.json"
	assert.Equal(t, expected, result)
}
