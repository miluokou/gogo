package poi

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"mvc/service"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	filesPerPage = 10
	cacheFile    = "startIndexCache.txt"
)

var fileCache = make(map[string]bool)
var cacheMutex = &sync.Mutex{}

func openFile(filePath string) (*os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func closeFile(file *os.File) {
	file.Close()
}

func isFileProcessed(filePath string) bool {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	return fileCache[filePath]
}

func markFileAsProcessed(filePath string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	fileCache[filePath] = true
}

func getCachedStartIndex() (int, error) {
	file, err := os.Open(cacheFile)
	if err != nil {
		return 0, fmt.Errorf("缓存文件打开失败: %s", err.Error())
	}
	defer file.Close()

	var startIndex int
	_, err = fmt.Fscanf(file, "%d\n", &startIndex)
	if err != nil {
		return 0, fmt.Errorf("缓存初始值读取失败: %s", err.Error())
	}

	return startIndex, nil
}

func setCachedStartIndex(startIndex int) error {
	file, err := os.Create(cacheFile)
	if err != nil {
		return fmt.Errorf("创建缓存文件失败: %s", err.Error())
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%d\n", startIndex)
	if err != nil {
		return fmt.Errorf("写入缓存文件失败: %s", err.Error())
	}

	return nil
}

func CsvToPoi(c *gin.Context) {
	dir := "public"
	absDir, _ := filepath.Abs(dir)
	csvFiles := make([]string, 0)

	_ = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".csv") {
			csvFiles = append(csvFiles, path)
		}
		return nil
	})

	var errors []string

	startIndex, _ := getCachedStartIndex()
	endIndex := startIndex + filesPerPage
	if endIndex > len(csvFiles) {
		endIndex = len(csvFiles)
	}

	for _, filePath := range csvFiles[startIndex:endIndex] {
		if isFileProcessed(filePath) {
			continue
		}
		markFileAsProcessed(filePath)

		file, _ := openFile(filePath)
		reader := csv.NewReader(file)
		records, _ := reader.ReadAll()
		closeFile(file)

		var data [][]string
		for _, record := range records {
			if len(record) > 0 && record[0] == "名称" {
				continue
			}
			data = append(data, record)
		}

		service.StoreData20231022("poi_2023_01", data)

		count := len(data)
		countStr := strconv.Itoa(count)
		fmt.Println(fmt.Sprintf("%s 文件存储完毕 %s 行", filePath, countStr))

		_ = os.Remove(filePath)
		markFileAsProcessed(filePath)

		time.Sleep(time.Duration(rand.Intn(2500)+1000) * time.Millisecond)
	}

	if len(errors) > 0 {
		c.JSON(500, gin.H{"error": strings.Join(errors, ", ")})
		return
	}

	setCachedStartIndex(endIndex) // 设置新的索引值

	c.JSON(200, gin.H{"message": "CSV to POI conversion completed"})
}
