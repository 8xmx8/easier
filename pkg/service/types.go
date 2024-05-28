package service

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// nolint
type ResponseList struct {
	Total int `json:"total"`
	List  any `json:"list"`
}

// Response 定义响应
// nolint
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func (resp *Response) JSON() gin.H {
	dataB, _ := json.Marshal(resp)
	h := new(gin.H)
	_ = json.Unmarshal(dataB, h)
	return *h
}

func ResponseOK(msg string, data any) *Response {
	return &Response{
		Code:    0,
		Message: msg,
		Data:    data,
	}
}

func ResponseError(msg string, data any) *Response {
	return &Response{
		Code:    -1,
		Message: msg,
		Data:    data,
	}
}
