package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"mvc/routers"
	"mvc/service"
	"mvc/utils"
	"time"
)

func main() {

	// 启动常驻的goroutine
	go backgroundRoutine()

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

func backgroundRoutine() {
	for {
		currentTime := time.Now()
		formattedTime := currentTime.Format("2006-01-02 15:04:05") // 使用指定的日期时间格式

		service.LogInfo("backgroundRoutine 正常执行中：" + formattedTime)

		time.Sleep(time.Minute * 2) // 可以添加适当的休眠时间，避免过于频繁地执行任务
	}
}
