package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

func ConvertToExcel(c *gin.Context) {
	imagePath := "test.png"     // 图像文件路径
	outputFile := "output.xlsx" // 输出的Excel文件路径

	// 使用tesseract-ocr进行OCR识别
	text, err := runTesseractOCR(imagePath, "chi_sim") // 使用简体中文语言包
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法执行OCR识别"})
		return
	}

	// 创建Excel文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建Excel工作表"})
		return
	}

	row := sheet.AddRow()
	cell := row.AddCell()

	// 将文本转换为UTF-8编码
	utf8Text, err := gbkToUtf8(text)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法转换文本编码"})
		return
	}

	cell.Value = utf8Text

	// 设置单元格字体为宋体
	font := xlsx.NewFont(12, "宋体")
	style := xlsx.NewStyle()
	style.Font = *font
	cell.SetStyle(style)

	err = file.Save(outputFile)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法保存Excel文件"})
		return
	}

	fmt.Println("Excel文件保存成功")

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

func gbkToUtf8(text string) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Reader := transform.NewReader(strings.NewReader(text), decoder)
	utf8Bytes, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		return "", err
	}
	utf8Text := string(utf8Bytes)
	return utf8Text, nil
}
