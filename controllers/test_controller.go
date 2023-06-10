// mvc/controllers/user_controller.go
package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"mvc/models"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DB struct {
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"db"`
}

func TestEnv(c *gin.Context) {
    // 读取配置文件中的 DB_USER 和 DB_PASS
    configData, err := ioutil.ReadFile("config.yaml")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading config file"})
        return
    }

    var config Config
    err = yaml.Unmarshal(configData, &config)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing config file"})
        return
    }

    fmt.Printf("DB_USER: %s, DB_PASS: %s\n", config.DB.User, config.DB.Pass)

    // 获取所有用户
    users, err := models.GetUsers()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, users)
}
