package service

import (
	"fmt"
	"log"
	"mvc/utils"
)

var Logger *log.Logger

func LogInfo(v ...interface{}) {
	message := fmt.Sprint(v...)
	//fmt.Println("[INFO]", message)
	utils.InitLogger()

	// 将utils.Logger赋值给全局的Logger变量
	Logger = utils.Logger

	// 使用日志记录器进行日志输出
	Logger.Println(message)
}
