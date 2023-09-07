package service

import (
	"fmt"
	"log"
	"mvc/utils"
	"runtime"
)

var Logger *log.Logger

func LogInfo(v ...interface{}) {
	message := fmt.Sprint(v...)

	// 获取调用日志函数的方法名和行号
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		message = fmt.Sprintf("[%s:%d] %s", funcName, line, message)
	}

	utils.InitLogger()

	// 将utils.Logger赋值给全局的Logger变量
	Logger = utils.Logger

	// 使用日志记录器进行日志输出
	Logger.Println(message)
}
