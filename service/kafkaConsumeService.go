package service

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
)

// KafkaConsumerService 定义 Kafka 消费者服务
type KafkaConsumerService struct {
	reader *kafka.Reader
}

// NewKafkaConsumerService 创建 Kafka 消费者服务实例
func NewKafkaConsumerService(brokers []string, topic string, groupID string) (*KafkaConsumerService, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &KafkaConsumerService{
		reader: reader,
	}, nil
}

// ConsumeMessage 从 Kafka 消费消息
func (s *KafkaConsumerService) ConsumeMessage(ctx context.Context) ([]byte, error) {
	msg, err := s.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to consume message: %w", err)
	}

	return msg.Value, nil
}

// Close 关闭 Kafka 消费者连接
func (s *KafkaConsumerService) Close() error {
	if s.reader != nil {
		return s.reader.Close()
	}
	return nil
}
