package video

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTextWatermark(t *testing.T) {
	watermark, err := TextWatermark(context.Background(),
		"font.ttf", "小杨", 40,
		800, 60, 10, 50, 72, 1001010001)
	assert.NoError(t, err)
	fmt.Println(watermark)
}
