package main

import (
	"github.com/gin-gonic/gin"
	"mvc/routers"
	"mvc/utils"
	"github.com/spf13/viper"
	"fmt"
)

func main() {

	// 初始化 viper 库
	viper.SetConfigName("config") // 配置文件名称（无扩展名）
	viper.AddConfigPath(".")      // 配置文件路径
	viper.SetConfigType("yaml")   // 如果配置文件的名称中没有扩展名，则需要配置此项
	err := viper.ReadInConfig()   // 查找并读取配置文件
	if err != nil {               // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// 创建 Gin 路由引擎
	router := gin.Default()

	// 连接数据库或执行其他初始化操作
	utils.InitDatabase()

	// 注册路由
	routers.SetupRouter(router)

	// 启动服务器，监听指定端口
	router.Run(":9090")
}
