package errorhandler

import (
	"mall/pkg/e"
	util "mall/pkg/utils"
	"mall/serializer"

	// "github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HandleError 是一个通用的错误处理函数
func HandleError(err error, code int, data interface{}) serializer.Response {
	// 检查 err 是否为 nil
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	} else {
		errorMsg = "Unknown error"
	}
	// 记录错误日志
	util.LogrusObj.WithFields(logrus.Fields{
		"error": err,
		"code":  code,
	}).Error("Error occurred") // 使用 Error 级别记录错误日志
	// 返回标准化的错误响应
	return serializer.Response{
		Status: code,
		Data:   data,
		Msg:    e.GetMsg(code),
		Error:  errorMsg,
	}
}
