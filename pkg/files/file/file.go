package file

import (
	"os"
)

// FolderExistOrCreate 检查目录是否存在, 如果不存在就创建
// 递归创建目录
func FolderExistOrCreate(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 目录不存在，创建目录
		err := os.MkdirAll(path, 0o755)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsExist 检查文件是存在的
func IsExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
