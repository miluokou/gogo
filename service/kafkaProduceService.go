package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
)

// ProduceMessage 发送消息到 Kafka 主题
func ProduceMessage(message string) {
	// 从请求获取消息内容

	LogInfo("准备发送消息到 Kafka")
	LogInfo(fmt.Sprintf("消息内容: %s", message))

	// 创建 Kafka 生产者
	producer, err := CreateProducer()
	if err != nil {
		LogInfo(fmt.Errorf("创建 Kafka 生产者失败：%w", err).Error())
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
		LogInfo(fmt.Errorf("发送消息失败：%w", err).Error())

		return
	}

	LogInfo("消息发送成功")
}

func CreateProducer() (*kafka.Writer, error) {
	// 配置 Kafka 代理地址
	brokers := []string{"localhost:9092"}

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

	LogInfo("准备发送消息到 Kafka")
	LogInfo(fmt.Sprintf("消息内容: %s", string(message.Value)))

	err := writer.WriteMessages(ctx, message)
	if err != nil {
		LogInfo(fmt.Errorf("发送消息失败：%w", err).Error())
	} else {
		LogInfo("消息发送成功")
	}

	return err
}
