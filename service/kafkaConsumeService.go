package service

import (
	"log"
	"time"

	"context"
	"github.com/segmentio/kafka-go"
)

type Message struct {
	Value string `json:"value"`
}

func ConsumeMessages() ([]Message, error) {
	// 创建 Kafka 消费者
	consumer, err := CreateConsumer()
	if err != nil {
		return nil, err
	}
	defer consumer.Close()

	// 消费消息
	messages, err := Consume(consumer)
	if err != nil {
		return nil, err
	}

	LogInfo(messages)
	return messages, nil
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
			log.Println("从 Kafka 读取消息失败:", err)
			break
		}

		message := Message{Value: string(m.Value)}
		messages = append(messages, message)
	}

	return messages, nil
}
