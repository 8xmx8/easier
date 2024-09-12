package video

import (
	"bytes"
	"context"
	"github.com/8xmx8/easier/pkg/files/file"
	"github.com/golang/freetype"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// TextWatermark 获取视频水印,fontFile为字体文件,水印的字体格式
// 短视频格式 fontSize: 40,imgW: 800,imgH: 60,textX: 10,textY: 50,DPI: 72
func TextWatermark(ctx context.Context, fontFile, text string,
	fontSize, imgW, imgH, textX, textY int, DPI float64, actorId int64) (string, error) {
	f, err := file.ParseFont(fontFile)
	if err != nil {
		return "", err
	}
	// 设置文本颜色
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 128}
	// 创建一个新的RGBA图片
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	// 将背景颜色设置为透明
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Transparent}, image.Point{}, draw.Src)
	// 创建一个新的freetype上下文
	c := freetype.NewContext()
	c.SetDPI(DPI)
	c.SetFont(f)
	c.SetFontSize(float64(fontSize))
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(textColor))
	// 在图片上绘制文本
	pt := freetype.Pt(textX, textY)
	_, err = c.DrawString(text, pt)
	if err != nil {
		return "", err
	}
	// 将图像保存到内存中
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", err
	}
	WatermarkPNGName := GenerateNameWatermark(actorId, text)
	// 将图片保存到文件
	_, err = file.Upload(ctx, WatermarkPNGName, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", err
	}
	return WatermarkPNGName, nil
}
