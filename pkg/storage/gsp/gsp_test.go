package gsp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/nfnt/resize"
	"github.com/stretchr/testify/assert"
	"image"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"testing"
)

const (
	accessKey = "I0SQQA7FBG2NLD0AN9ET"
	secretKey = "vVMB+ghZHACgb94fir5AiASYJhRnOmMVk4+BAFoK"
	regions   = "eyfm"
	addr      = "http://172.16.20.30:9000"
)

func TestGSP(t *testing.T) {
	g, err := NewGSP(addr, accessKey, secretKey, regions)
	assert.NoError(t, err)

	// 写入
	local, err := g.PutS3Object(context.Background(), "testimages", "212132", "application/octet-stream", []byte("hello world"))
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s/%s/%s", addr, "testimages", "212132"), local)

	// 读取
	data, err := g.GetS3Object(context.Background(), "testimages", "212132")
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), data)

	buf := aws.NewWriteAtBuffer([]byte{})
	err = g.GetS3ObjectWithWriter(context.Background(), "testimages", "212132", buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), buf.Bytes())
}

func TestImages(t *testing.T) {
	g, err := NewGSP(addr, accessKey, secretKey, regions)
	assert.NoError(t, err)

	imageBytes, err := readImageFile("./1389a42ec6a1c0d56085b23b0662bbb.png")
	assert.NoError(t, err)
	originalSize := len(imageBytes)

	compressedImageBytes, err := compressImage(imageBytes)
	assert.NoError(t, err)

	local, err := uploadCompressedImage(g, compressedImageBytes)
	assert.NoError(t, err)
	expectedURL := fmt.Sprintf("%s/%s/%s", addr, "testimages", "compressed_image.jpg")
	assert.Equal(t, expectedURL, local)

	downloadedImageBytes, err := downloadCompressedImage(g)
	assert.NoError(t, err)

	assertCompressionRatio(t, originalSize, downloadedImageBytes)

	err = saveDownloadedImage(downloadedImageBytes)
	assert.NoError(t, err)
}

func readImageFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func compressImage(imageBytes []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}

	resizedImg := resize.Resize(uint(img.Bounds().Dx()/2), 0, img, resize.Lanczos3)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resizedImg, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func uploadCompressedImage(g *GSP, compressedImageBytes []byte) (string, error) {
	return g.PutS3Object(context.Background(), "testimages", "compressed_image.jpg", "application/octet-stream", compressedImageBytes)
}

func downloadCompressedImage(g *GSP) ([]byte, error) {
	return g.GetS3Object(context.Background(), "testimages", "compressed_image.jpg")
}

func assertCompressionRatio(t *testing.T, originalSize int, downloadedImageBytes []byte) {
	storedSize := len(downloadedImageBytes)
	compressionRatio := calculateCompressionRatio(originalSize, storedSize)
	fmt.Println("Compression Ratio:", compressionRatio)
}

func calculateCompressionRatio(originalSize, storedSize int) float64 {
	if originalSize == 0 {
		return 0
	}
	return float64(storedSize) / float64(originalSize)
}

func saveDownloadedImage(downloadedImageBytes []byte) error {
	return ioutil.WriteFile("downloaded_compressed_image.jpg", downloadedImageBytes, 0644)
}

func CalculateCompressionRatio(originalSize, storedSize int) float64 {
	if originalSize == 0 {
		return 0
	}
	return float64(storedSize) / float64(originalSize)
}
