package kafka

import (
	"log"
	"net/http"
	"time"

	"context"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

type Message struct {
	Value string `json:"value"`
}

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
	brokers := []string{"localhost:9092"}
	topic := "test-data2"
	groupID := "test-consumer-group"
	partition := 0

	config := kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		Partition:   partition,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
		MaxWait:     time.Millisecond * 100, // 设置等待时间为 100 毫秒
	}
	reader := kafka.NewReader(config)

	return reader, nil
}

func Consume(reader *kafka.Reader) ([]Message, error) {
	var messages []Message

	for {
		// 从 Kafka 消费消息
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("Failed to read messag from Kafka:", err)
			break
		}
		message := Message{Value: string(m.Value)}
		messages = append(messages, message)

		// 收到消息后立即返回结果
		return messages, nil
	}

	return messages, nil
}
