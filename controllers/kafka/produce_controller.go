package kafka

import (
	"context"
	"errors"
	"fmt"
	"mvc/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

// ErrorResponse 定义错误响应结构体
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ProduceMessage 发送消息到 Kafka 主题
func ProduceMessage(c *gin.Context) {
	// 从请求获取消息内容
	message := c.PostForm("message")

	service.LogInfo("准备发送消息到 Kafka")
	service.LogInfo(fmt.Sprintf("消息内容: %s", message))

	// 创建 Kafka 生产者
	producer, err := CreateProducer()
	if err != nil {
		service.LogInfo(fmt.Errorf("创建 Kafka 生产者失败：%w", err).Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "创建 Kafka 生产者失败", Message: err.Error()})
		return
	}
	defer producer.Close()

	// 构建消息对象
	kafkaMsg := kafka.Message{
		Value: []byte(message),
	}

	// 发送消息
	err = SendMessage(producer, context.Background(), kafkaMsg)
	if err != nil {
		service.LogInfo(fmt.Errorf("发送消息失败：%w", err).Error())

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "发送消息失败", Message: err.Error()})
		return
	}

	service.LogInfo("消息发送成功")
	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully"})
}

func CreateProducer() (*kafka.Writer, error) {
	// 配置 Kafka 代理地址
	brokers := []string{"47.100.242.199:9092"}

	// 配置 Kafka 主题
	topic := "test-data2"

	// 创建 Kafka 生产者
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})

	return writer, nil
}

func SendMessage(writer *kafka.Writer, ctx context.Context, message kafka.Message) error {
	if writer == nil {
		return errors.New("kafka.Writer is nil")
	}

	service.LogInfo("准备发送消息到 Kafka")
	service.LogInfo(fmt.Sprintf("消息内容: %s", string(message.Value)))

	err := writer.WriteMessages(ctx, message)
	if err != nil {
		service.LogInfo(fmt.Errorf("发送消息失败：%w", err).Error())
	} else {
		service.LogInfo("消息发送成功")
	}

	return err
}
