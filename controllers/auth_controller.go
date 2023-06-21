// mvc/controllers/auth_controller.go

package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"errors"
	"mvc/models"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	
    fmt.Printf("getWxPhoneNumber 方法中的sessionkey: %s\n", sessionKey)
    fmt.Printf("getWxPhoneNumber 方法中的encryptedData: %s\n", encryptedData)
    fmt.Printf("getWxPhoneNumber 方法中的iv: %s\n", iv)
    fmt.Printf("getWxPhoneNumber 方法中的phoneCode: %s\n", phoneCode)
	// 构造请求体
	reqBody := map[string]string{
		"code":phoneCode,
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
    fmt.Printf("getWxPhoneNumber方法中请求微信接口返回的 response body: %s\n", string(body)) // 打印返回值
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
		errorMsg := fmt.Sprintf("failed to get session_key and openid from weixin: %v", err)
		fmt.Println(errorMsg) // 输出错误日志到终端
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to read response body: %v", err)
		fmt.Println(errorMsg) // 输出错误日志到终端
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}
	
	fmt.Printf("code: %s\n", req.Code)
	fmt.Printf("encryptedData: %s\n", req.EncryptedData)
	fmt.Printf("iv: %s\n", req.IV)
	fmt.Printf("phoneCode: %s\n", req.PhoneCode)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		errorMsg := fmt.Sprintf("failed to unmarshal json: %v", err)
		fmt.Println(errorMsg) // 输出错误日志到终端
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}

	// 检查是否成功获取 session_key 和 openid
	if _, ok := data["session_key"]; !ok {
		errorMsg := "failed to get session_key"
		fmt.Println(errorMsg) // 输出错误日志到终端
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}
		fmt.Printf("session_key是: %s\n", data["session_key"])
	if _, ok := data["openid"]; !ok {
		errorMsg := "failed to get openid"
		fmt.Println(errorMsg) // 输出错误日志到终端
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}
	fmt.Printf("openid是: %s\n", data["openid"])

	// 解密用户手机号码
	phoneNumber, err := getWxPhoneNumber(req.PhoneCode, req.EncryptedData, req.IV, data["session_key"].(string))
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
    //已经获取到了手机号
    _, err = models.GetUserByPhoneNumber(phoneNumber)
    if err != nil {
		fmt.Println(phoneNumber) // 输出错误日志到终端
		newUser := models.User{
            Name:       "John Doe",
            Email:      "johndoe@example.com",
            PhoneNumber: phoneNumber,
        }
    	_, err := models.CreateUser(newUser)

    	if err != nil {
    		errorMsg := "新增用户失败"
    		fmt.Println(errorMsg) // 输出错误日志到终端
    		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
    		return
    	}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
// 		return
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
    	errorMsg := fmt.Sprintf("failed to get phone number: %v", data)
    	fmt.Println(errorMsg) // 输出错误日志到终端
    	return "", fmt.Errorf(errorMsg)
    }

    
    return data["phoneNumber"].(string), nil
}