package middleware

import (
	"github.com/beego/beego/v2/server/web/context"
)

// CorsHandler 处理跨域请求
func CorsHandler(ctx *context.Context) {
	ctx.Output.Header("Access-Control-Allow-Origin", "*")
	ctx.Output.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	ctx.Output.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization")
	ctx.Output.Header("Access-Control-Allow-Credentials", "true")
	
	// 处理预检请求
	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.SetStatus(200)
		ctx.ResponseWriter.WriteHeader(200)
		return
	}
} 