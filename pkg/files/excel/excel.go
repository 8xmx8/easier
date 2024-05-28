package excel

import (
	"context"
	"github.com/xuri/excelize/v2"
	"mime/multipart"
	"strings"
	"sync"

	"github.com/8xmx8/easier/pkg/files/file"
	"github.com/8xmx8/easier/pkg/logger"
)

// nolint
type ExcelParser struct {
	excelize *excelize.File
	logg     logger.Logger
	mu       sync.Mutex // 添加一个互斥锁
}

type ParserWithOption func(*ExcelParser)

func WithLogger(logg logger.Logger) ParserWithOption {
	return func(ep *ExcelParser) {
		ep.logg = logg
	}
}

// NewExcelParser 创建excel的解析器
func NewExcelParser(ctx context.Context, filePath string, ops ...ParserWithOption) (*ExcelParser, error) {
	var ef *excelize.File
	isExist, err := file.IsExist(filePath)
	if err != nil {
		return nil, err
	}
	if isExist {
		// 存在, 说明要创建读取
		ef, err = excelize.OpenFile(filePath)
	} else {
		// 不存在说明需要创建写入
		ef = excelize.NewFile()
		ef.Path = filePath
	}
	if err != nil {
		return nil, err
	}
	ep := &ExcelParser{
		excelize: ef,
		logg:     logger.DefaultLogger(),
	}

	for _, op := range ops {
		op(ep)
	}
	return ep, nil
}

// Save 保存文件
func (ep *ExcelParser) Save(ctx context.Context) error {
	defer ep.excelize.Close()
	return ep.excelize.Save()
}

// rowsNum: 指定行数, 如果小于0, 则是所有行数据; 如果rowsNum>实际行数, 返回所有行数据
func (ep *ExcelParser) ReadRows(ctx context.Context, sheet string, rowsNum int) (dataList [][]string, err error) {
	rows, err := ep.excelize.Rows(sheet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rowsNum < 0 {
		// 返回所有数据
		return ep.excelize.GetRows(sheet)
	}
	dataList = make([][]string, 0, rowsNum)
	var curLine int
	for rows.Next() {
		if curLine >= rowsNum {
			return
		}
		curLine++
		row, colErr := rows.Columns()
		if colErr != nil {
			ep.logg.Error(logger.ErrorReadFile, "excel读取错误", logger.ErrorField(err), logger.MakeField("line_num", curLine))
			continue
		}
		dataList = append(dataList, row)
	}

	return
}

// WriteData 写入数据到指定的工作表
func (ep *ExcelParser) WriteRow(ctx context.Context, sheet string, row int, data []interface{}) error {
	ep.mu.Lock()         // 加锁
	defer ep.mu.Unlock() // 解锁（在函数返回时）
	for col, value := range data {
		cell, err := excelize.CoordinatesToCellName(col+1, row)
		if err != nil {
			return err
		}
		if err := ep.excelize.SetCellValue(sheet, cell, value); err != nil {
			return err
		}
	}
	return nil
}

// AsyncReadAllRows 异步读取指定sheet的所有数据
func (ep *ExcelParser) AsyncReadAllRows(ctx context.Context, sheet string) (<-chan []string, error) {
	rows, err := ep.excelize.Rows(sheet)
	if err != nil {
		return nil, err
	}
	dataChan := make(chan []string, 10)
	var curLine uint
	go func(ctx context.Context) {
		defer rows.Close()
		defer close(dataChan)
		for rows.Next() {
			curLine++
			row, colErr := rows.Columns()
			if colErr != nil {
				ep.logg.Error(logger.ErrorReadFile, "excel读取错误", logger.ErrorField(err), logger.MakeField("line_num", curLine))
				continue
			}
			dataChan <- row
		}
	}(ctx)
	return dataChan, err
}

func (ep *ExcelParser) GetFileObject() *excelize.File {
	return ep.excelize
}

func ParseTargetsFromExcel(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return "", err
	}

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return "", err
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return "", err
	}

	var targets []string
	for i, row := range rows {
		if i > 0 {
			if len(row) > 0 {
				targets = append(targets, row[0])
			}
		}
	}

	return strings.Join(targets, ","), nil
}
