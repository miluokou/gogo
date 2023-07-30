package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
	"mvc/routers"
	"mvc/utils"
)

var Logger *log.Logger

func main() {
	// 初始化日志记录器
	utils.InitLogger()

	// 将utils.Logger赋值给全局的Logger变量
	Logger = utils.Logger

	// 使用日志记录器进行日志输出
	Logger.Println("加载了日志全局组件")

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
