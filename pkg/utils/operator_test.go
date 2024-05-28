package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIf(t *testing.T) {
	assert.Equal(t, 2, If(true, 2, 1))
	assert.Equal(t, 1, If(false, 2, 1))
}

func BenchmarkIf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		If(true, 2, 1)
	}
}
