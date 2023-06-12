// mvc/controllers/auth_controller.go

package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type WeChatLoginRequest struct {
	Code          string `json:"code"`
	EncryptedData string `json:"encryptedData"`
	IV            string `json:"iv"`
	PhoneCode     string `json:"phoneCode"`
}

type WeChatLoginResponse struct {
	OpenID      string `json:"openid"`
	SessionKey  string `json:"session_key"`
	PhoneNumber string `json:"phone_number"`
	Error       string `json:"error"`
}

func getWxPhoneNumber(phoneCode, encryptedData, iv, sessionKey string) (string, error) {
	// 构造解密用户手机号码的 URL
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", sessionKey)

	// 构造请求体
	reqBody := map[string]string{
		"code":          phoneCode,
		"encryptedData": encryptedData,
		"iv":            iv,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 发送 HTTP POST 请求解密用户手机号码
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	// 检查是否成功获取用户手机号码
	if _, ok := data["phone_info"].(map[string]interface{})["phoneNumber"]; !ok {
		return "", fmt.Errorf("failed to get phone number: %v", data)
	}

	return data["phone_info"].(map[string]interface{})["phoneNumber"].(string), nil
}


func WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}


	// 获取微信登录的配置信息
	appID := viper.GetString("wechat.app_id")
	appSecret := viper.GetString("wechat.app_secret")

	// 构造获取 session_key 和 openid 的 URL
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appID, appSecret, req.Code)

	// 发送 HTTP 请求获取 session_key 和 openid
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否成功获取 session_key 和 openid
	if _, ok := data["session_key"]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get session_key"})
		return
	}
	if _, ok := data["openid"]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get openid"})
		return
	}

	// // 解密用户手机号码
	// phoneNumber, err := getUserPhoneNumber(req.EncryptedData, req.IV, data["session_key"].(string))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// 解密用户手机号码
	phoneNumber, err := getWxPhoneNumber(req.PhoneCode, req.EncryptedData, req.IV, data["session_key"].(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回登录结果
	c.JSON(http.StatusOK, WeChatLoginResponse{
		OpenID:      data["openid"].(string),
		SessionKey:  data["session_key"].(string),
		PhoneNumber: phoneNumber,
	})
}

func getUserPhoneNumber(encryptedData, iv, sessionKey string) (string, error) {
	// 构造解密用户手机号码的 URL
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/getphonenumber?access_token=%s", sessionKey)

	// 构造请求体
	reqBody := map[string]string{
		"encryptedData": encryptedData,
		"iv":            iv,
	}
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 发送 HTTP POST 请求解密用户手机号码
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	// 检查是否成功获取用户手机号码
	if _, ok := data["phoneNumber"]; !ok {
		return "", fmt.Errorf("failed to get phone number: %v", data)
	}

	return data["phoneNumber"].(string), nil
}