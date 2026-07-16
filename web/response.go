package web

import "github.com/gin-gonic/gin"

func SendResponse(ctx *gin.Context, response ResponseDefault) {
	ctx.JSON(response.Code, response)
}

func SendResponseData(ctx *gin.Context, code int, message string, data any) {
	SendResponse(ctx, ResponseDefault{Status: true, Code: code, Data: data, Message: message})
}
