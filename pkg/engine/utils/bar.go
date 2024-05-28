package utils

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var consolas *truetype.Font

func init() {
	dir, _ := os.Getwd()
	fontBytes, err := os.ReadFile(filepath.Join(dir, "font", "simhei.ttf")) // 替换为你的字体文件路径
	if err != nil {
		panic(err)
	}
	consolas, err = freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
}

// 添加地址栏到图片的正上方
func AddAddressBar(img image.Image, url string) (image.Image, error) {
	addressBarHeight := 36
	barColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	textColor := color.Black
	fontSize := 24

	bounds := img.Bounds()
	rect := image.Rect(0, -addressBarHeight, bounds.Dx(), bounds.Dy())
	newImg := image.NewRGBA(rect)

	// 绘制原始图片
	draw.Draw(newImg, bounds, img, image.Point{}, draw.Src)
	// 绘制地址栏
	draw.Draw(newImg, image.Rect(0, -addressBarHeight, bounds.Dx(), 0), &image.Uniform{barColor}, image.Point{}, draw.Src)
	// 计算文本的宽度
	textWidth := len(url) * fontSize / 2
	textX := (bounds.Dx() - textWidth) / 2
	textY := -addressBarHeight / 2
	if err := drawText(newImg, url, textColor, fontSize, textX, textY); err != nil {
		return nil, err
	}
	return newImg, nil
}
func drawText(img draw.Image, text string, color color.Color, size, x, y int) error {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(consolas)
	c.SetFontSize(float64(size))
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color))
	pt := freetype.Pt(x, y)
	_, err := c.DrawString(text, pt)
	if err != nil {
		return err
	}
	return nil
}
