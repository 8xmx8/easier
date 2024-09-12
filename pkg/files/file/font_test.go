package file

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseFont(t *testing.T) {
	f, err := ParseFont("font.ttf")
	assert.NoError(t, err)
	t.Log(f.FUnitsPerEm())
}
