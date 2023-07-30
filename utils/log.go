package utils

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

var Logger *log.Logger

func InitLogger() {
	var err error

	// 获取当前日期作为日志文件名
	logFileName := time.Now().Format("2006-01-02") + ".log"

	// 创建日志目录
	logDir := "log/"
	if err = os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	// 打开日志文件，如果不存在则创建；如果存在则追加写入
	logFilePath := filepath.Join(logDir, logFileName)
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个新的日志记录器，并设置输出目的地为日志文件
	Logger = log.New(file, "", log.Ldate|log.Ltime)
}
