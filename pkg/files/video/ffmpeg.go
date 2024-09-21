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
	"os/exec"
)

// TextWatermark 获取视频水印,fontFile为字体文件,水印的字体格式
// [视频左上方] 短视频格式 fontSize: 40,imgW: 800,imgH: 60,textX: 10,textY: 50,DPI: 72
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

// AddWatermarkToVideo 添加水印逻辑
func AddWatermarkToVideo(ctx context.Context, WatermarkPNGName, videoTitle, videoRawFileName string, videoId, actorId int64) error {
	FinalFileName := GenerateFinalVideoName(actorId, videoTitle, videoId)
	RawFilePath := file.GetLocalPath(ctx, videoRawFileName)
	WatermarkPath := file.GetLocalPath(ctx, WatermarkPNGName)
	cmdArgs := []string{
		"-i", RawFilePath,
		"-i", WatermarkPath,
		"-filter_complex", "[0:v][1:v]overlay=10:10",
		"-f", "matroska", "-",
	}
	cmd := exec.Command("ffmpeg", cmdArgs...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	// Execute the command
	err := cmd.Run()
	if err != nil {
		return err
	}
	// Write the captured stdout to a file
	_, err = file.Upload(ctx, FinalFileName, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	return nil
}

// ExtractVideoCover 提取视频封面
func ExtractVideoCover(ctx context.Context, rawFileName, coverFileName string) error {
	RawFilePath := file.GetLocalPath(ctx, rawFileName)
	cmdArgs := []string{
		"-i", RawFilePath, "-vframes", "1", "-an", "-f", "image2pipe", "-",
	}
	cmd := exec.Command("ffmpeg", cmdArgs...)
	// Create a bytes.Buffer to capture stdout
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return err
	}
	// buf.Bytes() now contains the image data. You can use it to write to a file or send it to an output stream.
	_, err = file.Upload(ctx, coverFileName, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	return nil
}
