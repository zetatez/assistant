package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess       = 0
	CodeInvalidParams = 10001
	CodeDatabaseError = 10002
	CodeThirdPartyErr = 10003
	CodeNotFound      = 10004
	CodeUnauthorized  = 10005
	CodeForbidden     = 10006
	CodeServerError   = 10000
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // 仅成功时返回
}

func Ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func Err(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
	})
}
