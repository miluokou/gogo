package service

import (
	"fmt"
	"log"
	"mvc/utils"
	"runtime"
	"sync"
)

var (
	Logger      *log.Logger
	loggerMutex sync.Mutex
)

func init() {
	utils.InitLogger()
	Logger = utils.Logger
}

func LogInfo(v ...interface{}) {
	message := fmt.Sprint(v...)

	// 获取调用日志函数的方法名和行号
	pc, file, line, ok := runtime.Caller(1)
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		message = fmt.Sprintf("[%s:%s:%d] %s", file, funcName, line, message)
	}

	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// 使用日志记录器进行日志输出
	Logger.Println(message)
}
