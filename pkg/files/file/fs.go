package file

import (
	"context"
	"io"
	"net/url"
	"os"
	"path"
)

//TODO: 适用于短视频平台尺寸的缩放,视频添加水印

const (
	StorageType     = "fs"
	SystemStartPath = "/tmp"
	SystemBaseUrl   = "http://localhost/"
)

// TODO: context 应当使用飞行记录器记录，如果要使用，请将 go version 更新到 22.1.0 以上版本
type FSStorage struct {
}

func (f FSStorage) GetLocalPath(ctx context.Context, fileName string) string {
	return path.Join(SystemStartPath, fileName)
}

func (f FSStorage) Upload(ctx context.Context, fileName string, content io.Reader) (output *PutObjectOutput, err error) {
	all, err := io.ReadAll(content)
	if err != nil {
		return nil, err
	}
	filePath := path.Join(SystemStartPath, fileName)
	dir := path.Dir(filePath)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(filePath, all, os.FileMode(0755))
	if err != nil {
		return nil, err
	}
	return &PutObjectOutput{}, nil
}

func (f FSStorage) GetLink(ctx context.Context, fileName string) (string, error) {
	return url.JoinPath(SystemBaseUrl, fileName)
}

func (f FSStorage) IsFileExist(ctx context.Context, fileName string) (bool, error) {
	filePath := path.Join(SystemStartPath, fileName)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
