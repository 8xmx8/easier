package excel

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

type ExcelRequest struct {
	FileName   string        `json:"fileName"`
	DataKeys   []interface{} `json:"dataKeys"`
	DataValues []byte        `json:"dataValues"`
}
type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func initRequest() *ExcelRequest {
	users := []*User{
		{
			Id:   1,
			Name: "张三",
			Age:  18,
		},
		{
			Id:   2,
			Name: "李四",
			Age:  19,
		},
		{
			Id:   3,
			Name: "王五",
			Age:  20,
		},
		{
			Id:   4,
			Name: "赵六",
			Age:  21,
		},
	}
	bytes, err := json.Marshal(&users)
	if err != nil {
		return nil
	}
	return &ExcelRequest{
		FileName:   "test.xlsx",
		DataKeys:   []interface{}{"id", "name", "age"},
		DataValues: bytes,
	}
}
func TestNewExcelParser(t *testing.T) {
	req := initRequest()
	ctx := context.Background()
	fields := req.DataKeys
	fileName := req.FileName
	data := req.DataValues
	ep, err := NewExcelParser(ctx, fileName)
	if err != nil {
		t.Fatal(err)
	}
	ep.WriteRow(ctx, "Sheet1", 1, fields)

	t.Log(data)
	var dataMap []map[string]interface{}
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		t.Fatal(err)
	}
	rowNo := 2
	rowDatas := make([][]interface{}, 0)
	for _, item := range dataMap {
		// 按照DataKeys的顺序提取值
		rowData := make([]interface{}, len(fields))
		for i, key := range fields {
			rowData[i] = item[key.(string)] // 将interface{}类型的key转换为string
		}
		rowDatas = append(rowDatas, rowData)
	}
	err = ep.WriteRows(ctx, "Sheet1", rowNo, rowDatas)
	if err != nil {
		t.Fatal(err)
	}
	err = ep.Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fs := ep.GetFileObject()
	buf := new(bytes.Buffer)
	_, err = fs.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}
	i := buf.Bytes()
	t.Log(i)

	//gsp, _ := gsp.NewGSP("", "", "", "")

	//gsp.PutS3ObjectWithReader(ctx, "testimages", "test.xlsx", "application/octet-stream",buf)
}
