// mvc/controllers/auth_controller.go

package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WeChatLoginRequest struct {
	Code string `json:"code"`
}

type WeChatLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
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

	resp := WeChatLoginResponse{
		OpenID:     wechatResp.OpenID,
		SessionKey: wechatResp.SessionKey,
	}
	c.JSON(http.StatusOK, resp)
}

func getWeChatLoginResponse(code string) (*WeChatLoginResponse, error) {
	appID := "wx31fbd515cacb4d88"
	appSecret := "ca6754d99b05fc0139e9dd2b2a49ecd5"
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
