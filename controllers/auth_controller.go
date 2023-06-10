// mvc/controllers/auth_controller.go

package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WeChatLoginRequest struct {
	Code          string `json:"code"`
	EncryptedData string `json:"encryptedData"`
	IV            string `json:"iv"`
}

type WeChatLoginResponse struct {
	OpenID      string `json:"openid"`
	SessionKey  string `json:"session_key"`
	PhoneNumber string `json:"phone_number"`
}

func WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 调用微信登录验证接口，获取 openid 和 session_key
	wechatResp, err := getWeChatLoginResponse(req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to login"})
		return
	}

	// 获取用户手机号
	phoneNumber, err := getUserPhoneNumber(wechatResp.SessionKey, req.EncryptedData, req.IV)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user phone number"})
		return
	}

	resp := WeChatLoginResponse{
		OpenID:      wechatResp.OpenID,
		SessionKey:  wechatResp.SessionKey,
		PhoneNumber: phoneNumber,
	}
	c.JSON(http.StatusOK, resp)
}

func getWeChatLoginResponse(code string) (*WeChatLoginResponse, error) {
	appID := "wx5507ea2a74d21f58"
	appSecret := "5e7573fece9ccdba12b70a6650197693"
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appID, appSecret, code)

	// 发起 HTTP 请求到微信服务器
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析 JSON 响应
	var wechatResp WeChatLoginResponse
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		return nil, err
	}

	return &wechatResp, nil
}

func getUserPhoneNumber(sessionKey string, encryptedData string, iv string) (string, error) {
    appID := "wx5507ea2a74d21f58"
    appSecret := "5e7573fece9ccdba12b70a6650197693"
    // 获取 access_token
    accessTokenURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
    accessTokenResponse, err := http.Get(accessTokenURL)
    if err != nil {
        return "", err
    }
    defer accessTokenResponse.Body.Close()

    accessTokenBody, err := ioutil.ReadAll(accessTokenResponse.Body)
    if err != nil {
        return "", err
    }

    accessTokenData := make(map[string]interface{})
    err = json.Unmarshal(accessTokenBody, &accessTokenData)
    if err != nil {
        return "", err
    }

    accessToken := accessTokenData["access_token"].(string)
    fmt.Println("Access Token:", accessToken)

    // 获取用户手机号
    phoneNumberData := map[string]interface{}{
        "code":          sessionKey,
        "encryptedData": encryptedData,
        "iv":            iv,
    }

    requestBody, err := json.Marshal(phoneNumberData)
    if err != nil {
        return "", err
    }

    fmt.Println("Request Body:", string(requestBody))

    response, err := http.Post(fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", accessToken), "application/json", bytes.NewBuffer(requestBody))
    if err != nil {
        return "", err
    }
    defer response.Body.Close()

    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return "", err
    }

    fmt.Println("Response Body:", string(responseBody))

    responseData := make(map[string]interface{})
    err = json.Unmarshal(responseBody, &responseData)
    if err != nil {
        return "", err
    }

    if phoneNumberInfo, ok := responseData["phone_info"].(map[string]interface{}); ok {
        if phoneNumber, ok := phoneNumberInfo["phoneNumber"].(string); ok {
            return phoneNumber, nil
        }
    }

    return "", nil
}
