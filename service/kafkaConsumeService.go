package service

import (
	"encoding/json"
	"log"
	"time"

	"context"
	"github.com/segmentio/kafka-go"
)

type Message struct {
	Value string `json:"value"`
}

func ConsumeMessages() ([]PropertyData, error) {
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

	// 转换消息到 PropertyData 结构
	var propertyDataList []PropertyData
	for _, message := range messages {
		propertyData, err := convertToPropertyData(message)
		if err != nil {
			// 处理转换错误
			continue
		}
		propertyDataList = append(propertyDataList, propertyData)
	}

	LogInfo(propertyDataList)
	return propertyDataList, nil
}

func convertToPropertyData(message Message) (PropertyData, error) {
	// 解析 JSON 字符串到 PropertyData 结构
	var propertyData PropertyData
	err := json.Unmarshal([]byte(message.Value), &propertyData)
	if err != nil {
		return PropertyData{}, err
	}

	return propertyData, nil
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
