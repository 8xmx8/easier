package file

import (
	"context"
	"github.com/8xmx8/easier/pkg/files/video"
	"io"
	"net/url"
	"os"
	"path"
)

type FSStorage struct {
}

func (f FSStorage) GetLocalPath(ctx context.Context, fileName string) string {
	return path.Join(video.FileSystemStartPath, fileName)
}

func (f FSStorage) Upload(ctx context.Context, fileName string, content io.Reader) (output *PutObjectOutput, err error) {
	all, err := io.ReadAll(content)
	if err != nil {
		return nil, err
	}
	filePath := path.Join(video.FileSystemStartPath, fileName)
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
	return url.JoinPath(video.FileSystemBaseUrl, fileName)
}

func (f FSStorage) IsFileExist(ctx context.Context, fileName string) (bool, error) {
	filePath := path.Join(video.FileSystemStartPath, fileName)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
