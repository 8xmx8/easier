package webengine

import (
	"context"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScreenshot(t *testing.T) {
	ctx := context.Background()
	brow, err := NewBrowser(ctx)
	assert.NoError(t, err)
	defer brow.Close()
	subCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	page, err := brow.LoadPage(subCtx, "https://www.baidu.com/")
	assert.NoError(t, err)
	defer page.Close()
	title := page.GetTitle()
	assert.Equal(t, "百度一下，你就知道", title)
	html, err := page.GetHTML()
	assert.NoError(t, err)
	assert.Greater(t, len(html), 32)

	shot, err := page.ScreenshotPage(true, true)
	assert.NoError(t, err)
	outputFile, err := os.Create("my.png")
	assert.NoError(t, err)
	defer outputFile.Close()
	err = png.Encode(outputFile, shot)
	assert.NoError(t, err)
	urlPath, iconB, err := page.GetIcon()
	assert.NoError(t, err)
	assert.Equal(t, "https://www.baidu.com/favicon.ico", urlPath)
	iconFile, err := os.Create("favicon.ico")
	assert.NoError(t, err)
	defer iconFile.Close()
	_, err = iconFile.Write(iconB)
	assert.NoError(t, err)
}
