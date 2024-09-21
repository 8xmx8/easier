package video

import (
	"bytes"
	"context"
	"github.com/8xmx8/easier/pkg/files/file"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
	"time"
)

// TODO: args设计选择模式,cmd Run 为二级函数,根据key选择arg或者传入执行体
func TestAddWatermarkToVideo(t *testing.T) {
	//读取文件
	ctx := context.Background()

	actorId := int64(1001010001)
	title := "小杨的第一个作品"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	videoId := int64(r.Uint64())

	data, _ := os.ReadFile("0c6d1a503eba2f9a9915c2c3b2194d1d.mp4")
	reader := bytes.NewReader(data)

	fileName := GenerateRawVideoName(actorId, title, videoId)
	coverName := GenerateCoverName(actorId, title, videoId)
	_, err := file.Upload(ctx, fileName, reader)

	err = ExtractVideoCover(ctx, fileName, coverName)
	assert.NoError(t, err)

	watermark, err := TextWatermark(ctx,
		"font.ttf", "小杨", 40,
		800, 60, 10, 50, 72, actorId)
	assert.NoError(t, err)
	err = AddWatermarkToVideo(ctx, watermark, title, fileName, videoId, actorId)
	assert.NoError(t, err)
}
