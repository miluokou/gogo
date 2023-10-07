package ocr

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func ConvertToCSV(c *gin.Context) {
	imagePath := "test.png"    // 图像文件路径
	outputFile := "output.csv" // 输出的CSV文件路径

	// 使用tesseract-ocr进行OCR识别
	text, err := runTesseractOCR(imagePath, "chi_sim") // 使用简体中文语言包
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法执行OCR识别"})
		return
	}

	// 将文本转换为UTF-8编码
	utf8Text, err := convertEncoding(text)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法转换文本编码"})
		return
	}

	// 解析文本为CSV数据
	data := parseTextToCSV(utf8Text)

	// 创建CSV文件并写入数据
	err = writeDataToCSV(outputFile, data)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法保存CSV文件"})
		return
	}

	fmt.Println("CSV文件保存成功")

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "转换成功"})
}

func runTesseractOCR(imagePath string, language string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout", "--dpi", "300", "-l", language) // 使用指定语言包
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func convertEncoding(text string) (string, error) {
	reader := strings.NewReader(text)
	utf8Reader, err := charset.NewReader(reader, "")
	if err != nil {
		return "", err
	}

	utf8Bytes, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		return "", err
	}

	utf8Text := string(utf8Bytes)
	return utf8Text, nil
}

func parseTextToCSV(text string) [][]string {
	lines := strings.Split(text, "\n")
	data := make([][]string, len(lines))
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		data[i] = fields
	}
	return data
}

func writeDataToCSV(filename string, data [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
