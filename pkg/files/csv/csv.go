package csv

import (
	"context"
	"encoding/csv"
	"io"
	"os"

	"github.com/8xmx8/easier/pkg/logger"
)

// nolint
type CsvParser struct {
	logg logger.Logger
}

type ParserWithOption func(*CsvParser)

func WithNewLogger(logg logger.Logger) ParserWithOption {
	return func(cp *CsvParser) {
		cp.logg = logg
	}
}

func NewCsvParser(context context.Context, opts ...ParserWithOption) *CsvParser {
	cp := &CsvParser{
		logg: logger.DefaultLogger(),
	}
	for _, op := range opts {
		op(cp)
	}
	return cp
}

type ReaderWithOption func(*csv.Reader)

// ReaderWithComma 设置分隔符
// 默认 ','
func ReaderWithComma(flag rune) ReaderWithOption {
	return func(r *csv.Reader) {
		r.Comma = flag
	}
}

// ReaderWithComment 设置评论标识
// 默认是 空的
func ReaderWithComment(flag rune) ReaderWithOption {
	return func(r *csv.Reader) {
		r.Comment = flag
	}
}

// 读取整个csv文件
func (c *CsvParser) ReadFromFile(filepath string, ops ...ReaderWithOption) (dataList [][]string, err error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)
	for _, op := range ops {
		op(reader)
	}
	return reader.ReadAll()
}

// 异步读取csv文件
func (c *CsvParser) AsyncReadAllRows(ctx context.Context, filepath string, ops ...ReaderWithOption) (<-chan []string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	dataChan := make(chan []string, 10)
	reader := csv.NewReader(file)
	for _, op := range ops {
		op(reader)
	}
	go func(ctx context.Context) {
		defer file.Close()
		defer close(dataChan)
		for {
			var record []string
			record, err = reader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				c.logg.Error(logger.ErrorReadFile, "csv读取错误", logger.ErrorField(err))
				// TODO: @zcf 读取报错, 考虑记录错误信息
				continue
			}
			dataChan <- record
		}
	}(ctx)
	return dataChan, err
}
