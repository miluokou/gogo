package ocr

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 发起HTTP POST请求，并获取REST请求的结果
func requestPost(url string, param string) (string, error) {
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(param))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func processImage(c *gin.Context) {
	token := "[调用鉴权接口获取的token]"
	url := "https://aip.baidubce.com/rest/2.0/ocr/v1/table?access_token=" + token

	file, err := c.FormFile("image") // 从表单中获取上传的文件
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}

	// 读取文件内容并进行Base64编码
	imgData, err := ioutil.ReadFile(file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	imgBase64 := base64.StdEncoding.EncodeToString(imgData)

	body := fmt.Sprintf("image=%s", imgBase64)
	res, err := requestPost(url, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "HTTP request failed"})
		return
	}

	c.String(http.StatusOK, res)
}
