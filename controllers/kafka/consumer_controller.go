package kafka

import (
	"log"
	"net/http"

	"context"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

// ConsumeMessages 从 Kafka 主题消费消息
func ConsumeMessages(c *gin.Context) {
	// 创建 Kafka 消费者
	consumer, err := CreateConsumer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer consumer.Close()

	// 消费消息
	messages, err := Consume(consumer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func CreateConsumer() (*kafka.Reader, error) {
	// 配置 Kafka 代理地址
	brokers := []string{"localhost:9092"}

	// 配置 Kafka 主题、消费者组和分区
	topic := "test-data2"
	groupID := "test-consumer-group"
	partition := 0

	// 创建 Kafka 消费者
	config := kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     topic,
		GroupID:   groupID,
		Partition: partition,
		MinBytes:  10e3,
		MaxBytes:  10e6,
	}
	reader := kafka.NewReader(config)

	return reader, nil
}

func Consume(reader *kafka.Reader) ([]string, error) {
	var messages []string

	for {
		// 从 Kafka 消费消息
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("Failed to read message from Kafka:", err)
			break
		}
		messages = append(messages, string(m.Value))
	}

	return messages, nil
}
