package controllers

import (
	"net/http"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

// GetTopics 获取 Kafka 主题列表
func GetTopics(c *gin.Context) {
	// 配置 Kafka 客户端
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0 // 设置 Kafka 版本号

	// 连接到 Kafka 服务器
	client, err := sarama.NewClient([]string{"47.100.242.199:9092"}, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer client.Close()

	// 获取主题列表
	topics, err := client.Topics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"topics": topics})
}
