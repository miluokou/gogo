package csv

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"path/filepath"
)

// SplitFiles 分隔CSV文件并将分割好的文件移动到目标目录
func SplitFiles(c *gin.Context) {
	sourceDir := "public/dataSource" // 相对路径，根据实际情况修改

	fileList, err := getFiles(sourceDir)
	if err != nil {
		log.Fatal(err)
	}

	destDir := "public/done"

	for _, file := range fileList {
		err = splitFile(file, destDir)
		if err != nil {
			log.Printf("Failed to split file %s: %v", file, err)
		} else {
			err = deleteFile(file)
			if err != nil {
				log.Printf("Failed to delete file %s: %v", file, err)
			}
		}
	}
}

// getFiles 获取指定目录下的文件列表
func getFiles(dirPath string) ([]string, error) {
	var fileList []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get files from directory: %v", err)
	}

	return fileList, nil
}

// splitFile 分割CSV文件
func splitFile(filePath string, destDir string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV records: %v", err)
	}

	chunks := chunkRecords(records, 1000)

	for i, chunk := range chunks {
		newFilePath := filepath.Join(destDir, fmt.Sprintf("%s_%d.csv", filepath.Base(filePath), i+1))
		newFile, err := os.Create(newFilePath)
		if err != nil {
			return fmt.Errorf("failed to create new file: %v", err)
		}
		defer newFile.Close()

		writer := csv.NewWriter(newFile)
		err = writer.WriteAll(chunk)
		if err != nil {
			return fmt.Errorf("failed to write CSV records to new file: %v", err)
		}

		writer.Flush()
	}

	return nil
}

// chunkRecords 将记录按给定大小进行分割为多个块
func chunkRecords(records [][]string, size int) [][][]string {
	var chunks [][][]string

	for i := 0; i < len(records); i += size {
		end := i + size
		if end > len(records) {
			end = len(records)
		}
		chunk := make([][]string, end-i)
		copy(chunk, records[i:end])
		chunks = append(chunks, chunk)
	}

	return chunks
}

// deleteFile 删除文件
func deleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}
