package video

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

// TODO: sha256 生成hash字符串长度过大,考虑截取前16位，或者使用md5.Sum(s) [16]byte

// GenerateRawVideoName 生成初始视频名称，此链接仅用于内部使用，暴露给用户的视频名称
func GenerateRawVideoName(actorId int64, title string, videoId int64) string {
	hash := sha256.Sum256([]byte("RAW" + strconv.FormatInt(actorId, 10) + title + strconv.FormatInt(videoId, 10)))
	return hex.EncodeToString(hash[:]) + ".mp4"
}

// GenerateFinalVideoName 最终暴露给用户的视频名称
func GenerateFinalVideoName(actorId int64, title string, videoId int64) string {
	hash := sha256.Sum256([]byte(strconv.FormatInt(actorId, 10) + title + strconv.FormatInt(videoId, 10)))
	return hex.EncodeToString(hash[:]) + ".mp4"
}

// GenerateCoverName 生成视频封面名称
func GenerateCoverName(actorId int64, title string, videoId int64) string {
	hash := sha256.Sum256([]byte(strconv.FormatInt(actorId, 10) + title + strconv.FormatInt(videoId, 10)))
	return hex.EncodeToString(hash[:]) + ".png"
}

// GenerateAudioName 生成音频链接，此链接仅用于内部使用，不暴露给用户
func GenerateAudioName(videoFileName string) string {
	hash := sha256.Sum256([]byte("AUDIO_" + videoFileName))
	return hex.EncodeToString(hash[:]) + ".mp3"
}

// GenerateNameWatermark 生成用户名水印图片
func GenerateNameWatermark(actorId int64, Name string) string {
	hash := sha256.Sum256([]byte("Watermark" + strconv.FormatUint(uint64(actorId), 10) + Name))
	return hex.EncodeToString(hash[:]) + ".png"
}
