package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// KafkaProducerService 定义 Kafka 生产者服务
type KafkaProducerService struct {
	writer *kafka.Writer
}

// NewKafkaProducerService 创建 Kafka 生产者服务实例
func NewKafkaProducerService(brokers []string, topic string) (*KafkaProducerService, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})

	return &KafkaProducerService{
		writer: writer,
	}, nil
}

// SendMessage 发送消息到 Kafka
func (s *KafkaProducerService) SendMessage(message string) error {
	kafkaMsg := kafka.Message{
		Value: []byte(message),
	}

	err := s.writer.WriteMessages(context.Background(), kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	stats := s.writer.Stats()
	if stats.Writes == 0 {
		return errors.New("failed to send message to Kafka")
	}

	return nil
}

// Close 关闭 Kafka 生产者连接
func (s *KafkaProducerService) Close() error {
	if s.writer != nil {
		return s.writer.Close()
	}
	return nil
}
