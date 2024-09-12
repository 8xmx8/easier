package file

import (
	"github.com/golang/freetype/truetype"
	"os"
)

func ParseFont(fontFile string) (f *truetype.Font, err error) {
	// 加载字体文件
	fontBytes, err := os.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}
	// 解析字体文件
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	return font, err
}
