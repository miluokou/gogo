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

// CsvToPoi 改造后的函数，使用信号量控制并发访问共享资源
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

	concurrencyLimit := 5
	sem := make(chan struct{}, concurrencyLimit) // 创建信号量，限制并发访问数量

	var wg sync.WaitGroup

	for dirPath, fileNames := range csvFiles {
		for _, fileName := range fileNames {
			wg.Add(1) // 增加等待组计数器

			go func(dirPath, fileName string) {
				defer wg.Done()

				sem <- struct{}{} // 获取信号量

				service.LogInfo("开始读取文件")
				service.LogInfo(fileName)

				filePath := filepath.Join(dirPath, fileName)
				file, err := os.Open(filePath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s 打开失败: %s", fileName, err.Error()))
					return
				}
				reader := csv.NewReader(file)
				records, err := reader.ReadAll()
				file.Close()
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s 读取失败: %s", fileName, err.Error()))
					return
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
					errors = append(errors, fmt.Sprintf("%s 存储失败: %s", fileName, err.Error()))
					return
				}

				count := len(data)
				countStr := strconv.Itoa(count)
				service.LogInfo(fmt.Sprintf("%s 文件存储完毕 %s 行", fileName, countStr))

				err = os.Remove(filePath)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s 删除失败: %s", filePath, err.Error()))
				}
				time.Sleep(time.Duration(rand.Intn(2500)+1000) * time.Millisecond)

				<-sem // 释放信号量
			}(dirPath, fileName)
		}
	}

	wg.Wait() // 等待所有任务完成

	if len(errors) > 0 {
		c.JSON(500, gin.H{"errors": errors})
		return
	}

	c.JSON(200, csvFiles)
}
