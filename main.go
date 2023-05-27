package main

import (
	"github.com/gin-gonic/gin"
	"mvc/routers"
	"mvc/utils"
)

func main() {
	// 创建 Gin 路由引擎
	router := gin.Default()

	// 连接数据库或执行其他初始化操作
	utils.InitDatabase()

	// 注册路由
	routers.SetupRouter()

   

	// 启动服务器，监听指定端口
	router.Run(":8080")
}
