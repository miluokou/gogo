// mvc/controllers/auth_controller.go

package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// 获取微信登录的配置信息
func TestEnv(c *gin.Context) {
	appID := viper.GetString("wechat.app_id")
	appSecret := viper.GetString("wechat.app_secret")

	fmt.Printf("appID: %s\n", appID)
	fmt.Printf("appSecret: %s\n", appSecret)

	c.String(http.StatusOK, "appID: %s\nappSecret: %s\n", appID, appSecret)
}