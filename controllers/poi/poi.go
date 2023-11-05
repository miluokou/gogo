package poi

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math/rand"
	"mvc/service"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func CsvToPoi(c *gin.Context) {
	relativeDir := "public" // 相对路径，根据实际情况修改
	absDir, err := filepath.Abs(relativeDir)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	csvFiles := make(map[string][]string)
	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() { // 只处理目录
			files, err := ioutil.ReadDir(path)
			if err != nil {
				return err
			}
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".csv") { // 只处理CSV文件
					csvFiles[path] = append(csvFiles[path], file.Name())
				}
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var errors []string
	semaphore := make(chan struct{}, 1) // 控制文件处理的并发数量
	var mutex sync.Mutex

	for dirPath, fileNames := range csvFiles {
		for _, fileName := range fileNames {

			service.LogInfo("开始读取文件")
			service.LogInfo(fileName)

			filePath := filepath.Join(dirPath, fileName)
			semaphore <- struct{}{} // 获取信号量，控制并发数量
			go func(filePath string) {
				defer func() {
					<-semaphore // 释放信号量
				}()
				mutex.Lock()
				err := processFile(filePath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s 处理失败: %s", fileName, err.Error()))
					mutex.Unlock()
					return
				}
				mutex.Unlock()

				err = os.Remove(filePath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s 删除失败: %s", filePath, err.Error()))
				}
			}(filePath)
		}
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}

	if len(errors) > 0 {
		c.JSON(500, gin.H{"errors": errors})
		return
	}

	c.JSON(200, csvFiles)
}

func processFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	var data [][]string
	for _, record := range records {
		if len(record) > 0 && record[0] == "名称" {
			continue
		}

		data = append(data, record)
	}

	service.LogInfo("读取完文件组合成一个大data之后向es中存储")
	err = service.StoreData20231022("poi_2023_01", data)
	if err != nil {
		return err
	}

	count := len(data)
	countStr := strconv.Itoa(count)
	service.LogInfo(fmt.Sprintf("%s 文件存储完毕 %s 行", filePath, countStr))

	time.Sleep(time.Duration(rand.Intn(2500)+1000) * time.Millisecond)

	return nil
}
