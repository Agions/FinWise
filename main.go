package main

import (
	_ "blog/routers"
	"blog/models"
	"blog/middleware"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/core/logs"
)

func main() {
	// 初始化数据库
	models.InitDB()
	
	// 日志设置
	logs.SetLogger(logs.AdapterFile, `{"filename":"logs/finwise.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10}`)
	logs.Async()
	
	// 添加中间件
	beego.InsertFilter("/*", beego.BeforeRouter, middleware.CorsHandler)
	beego.InsertFilter("/api/*", beego.BeforeRouter, middleware.RateLimiter)
	beego.InsertFilter("/api/*", beego.BeforeRouter, middleware.JwtFilter)
	
	// 启动服务器
	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	beego.Run()
}
