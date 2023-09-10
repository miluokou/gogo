// mvc/controllers/auth_controller.go

package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"io/ioutil"
	"mvc/models"
	"mvc/service"
	"net/http"
	"sync"
	"time"
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
	Token       string `json:"token"`
}

var (
	wxAccessTokenCache     string
	wxAccessTokenExpiresAt time.Time
	wxAccessTokenMutex     sync.Mutex
)

func getStableAccessToken() string {
	wxAccessTokenMutex.Lock()
	defer wxAccessTokenMutex.Unlock()

	if wxAccessTokenCache != "" && wxAccessTokenExpiresAt.After(time.Now()) {
		return wxAccessTokenCache
	}

	// 获取微信登录的配置信息
	appID := viper.GetString("wechat.app_id")
	appSecret := viper.GetString("wechat.app_secret")

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("getStableAccessToken: failed to get access token: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("getStableAccessToken: failed to read response body: %v", err)
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("getStableAccessToken: failed to parse response body: %v", err)
		return ""
	}

	token, ok := data["access_token"].(string)
	if !ok {
		fmt.Printf("getStableAccessToken: invalid response: %s", string(body))
		return ""
	}

	expiresIn, ok := data["expires_in"].(float64)
	if !ok {
		fmt.Printf("getStableAccessToken: invalid response: %s", string(body))
		return ""
	}

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	wxAccessTokenCache = token
	wxAccessTokenExpiresAt = expiresAt

	return wxAccessTokenCache
}

func getWxPhoneNumber(phoneCode, encryptedData, iv, sessionKey string) (string, error) {
	// 构造解密用户手机号码的 URL
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", getStableAccessToken())
	// 构造请求体
	reqBody := map[string]string{
		"code": phoneCode,
		// 		"encryptedData": encryptedData,
		// 		"iv":            iv,
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
	// fmt.Printf("getWxPhoneNumber方法中请求微信接口返回的 response body: %s\n", string(body)) // 打印返回值
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	// 获取用户手机号码
	if phoneInfo, ok := data["phone_info"].(map[string]interface{}); ok {
		if phoneNumber, ok := phoneInfo["phoneNumber"].(string); ok {
			return phoneNumber, nil
		}
	}

	return "", errors.New("failed to get phone number")
}

func WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, "参数解析失败："+err.Error(), http.StatusBadRequest)
		return
	}

	appID := viper.GetString("wechat.app_id")
	appSecret := viper.GetString("wechat.app_secret")

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appID, appSecret, req.Code)

	resp, err := http.Get(url)
	if err != nil {
		service.LogInfo("无法获取微信登录信息：" + err.Error())
		handleError(c, "无法获取微信登录信息："+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	service.LogInfo("WeChatLogin")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		handleError(c, "读取响应内容失败："+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("微信返回的信息-----：%s\n", string(body))
	service.LogInfo("WeChatLogin body")
	service.LogInfo(body)
	var data struct {
		SessionKey string `json:"session_key"`
		OpenID     string `json:"openid"`
	}
	service.LogInfo("WeChatLogin")
	if err := json.Unmarshal(body, &data); err != nil {
		handleError(c, "解析JSON失败："+err.Error(), http.StatusInternalServerError)
		return
	}
	service.LogInfo("WeChatLogin")
	service.LogInfo(data)
	if data.SessionKey == "" {
		handleError(c, "无法获取session_key", http.StatusInternalServerError)
		return
	}
	service.LogInfo("WeChatLogin")
	fmt.Printf("session_key是：%s\n", data.SessionKey)
	if data.OpenID == "" {
		handleError(c, "无法获取openid", http.StatusInternalServerError)
		return
	}
	service.LogInfo("WeChatLogin")
	fmt.Printf("openid是：%s\n", data.OpenID)

	phoneNumber, err := getWxPhoneNumber(req.PhoneCode, req.EncryptedData, req.IV, data.SessionKey)
	if err != nil {
		handleError(c, "解密用户手机号码失败："+err.Error(), http.StatusInternalServerError)
		return
	}
	service.LogInfo("WeChatLogin")
	haveSaveUser, err := models.GetUserByPhoneNumber(phoneNumber)
	if err == nil {
		if haveSaveUser.ID != 0 {
			if haveSaveUser.OpenId != data.OpenID {
				haveSaveUser.OpenId = data.OpenID
				if _, err := models.UpdateUser(haveSaveUser); err != nil {
					fmt.Printf("Error: %s\n", err.Error())
					handleError(c, "openId或其他字段更新失败", http.StatusInternalServerError)
					return
				}
			}
		} else {
			newUser := models.User{
				Name:        "John Doe",
				Email:       "johndoe@example.com",
				PhoneNumber: phoneNumber,
				OpenId:      data.OpenID,
			}
			if _, err := models.CreateUser(newUser); err != nil {
				handleError(c, "新增用户失败", http.StatusInternalServerError)
				return
			}
		}
	} else {
		handleError(c, "查询用户的时候异常了", http.StatusInternalServerError)
		return
	}
	service.LogInfo("WeChatLogin")
	secretKey := "your_secret_key"
	salt := "your_salt"
	service.LogInfo("WeChatLogin")
	jwtController, err := service.NewJWTController(secretKey, salt)
	if err != nil {
		handleError(c, "创建JWTController失败："+err.Error(), http.StatusInternalServerError)
		return
	}
	service.LogInfo("WeChatLogin")
	token, err := jwtController.GenerateToken(haveSaveUser.ID, time.Hour*24*7) // 假设用户ID为123，有效期为24小时
	if err != nil {
		handleError(c, "生成令牌失败："+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("生成的token是hfudihfsidh：", token)
	fmt.Println("data.OpenID 的值为：", data.OpenID)
	service.LogInfo("WeChatLogin")
	c.JSON(http.StatusOK, WeChatLoginResponse{
		OpenID:      data.OpenID,
		SessionKey:  data.SessionKey,
		PhoneNumber: phoneNumber,
		Token:       token,
	})
}

func handleError(c *gin.Context, errorMsg string, statusCode int) {
	fmt.Println(errorMsg) // 输出错误日志到终端
	c.JSON(statusCode, gin.H{"error": errorMsg})
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
		errorMsg := fmt.Sprintf("failed to get phone number: %v", data)
		fmt.Println(errorMsg) // 输出错误日志到终端
		return "", fmt.Errorf(errorMsg)
	}

	return data["phoneNumber"].(string), nil
}
