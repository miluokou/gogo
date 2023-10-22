package poi

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mvc/service"
	"os"
	"path/filepath"
	"strings"
)

/**
* 把购买的poi的csv文件存储到es中
 */

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
	for dirPath, fileNames := range csvFiles {
		for _, fileName := range fileNames {
			service.LogInfo("文件名是" + fileName)
			filePath := filepath.Join(dirPath, fileName)
			file, err := os.Open(filePath)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			defer file.Close()

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			// Process the CSV records as needed
			// For example, you can access each record using a loop:
			var data [][]string
			for _, record := range records {
				if len(record) > 0 && record[0] == "名称" {
					// 第一列元素与目标字符串相等
					// 在这里处理逻辑
					continue
				}

				data = append(data, record)
			}

			err = service.StoreData20231022("poi_2023_01", data)
			if err != nil {
				//service.LogInfo(result)
				fmt.Printf("Failed to store data in Elasticsearch: %v", err)
				c.Set("error", fileName+"无法存储数据到Elasticsearch")
				//return
			}
			service.LogInfo(fileName + "文件存储完毕")
			break
		}
		break
	}

	c.JSON(200, csvFiles)
}
