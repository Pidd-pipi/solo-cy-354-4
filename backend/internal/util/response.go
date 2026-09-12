package util

import "github.com/gin-gonic/gin"

// Response is the unified API envelope {code,message,data}.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// OK writes a success response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Response{Code: 0, Message: "ok", Data: data})
}

// Fail writes an error response with the given HTTP status and business code.
func Fail(c *gin.Context, status, code int, message string) {
	c.JSON(status, Response{Code: code, Message: message, Data: nil})
}
