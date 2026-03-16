package kafkaproducer

import (
	"context"
	"encoding/json"
	"log/slog"
	"news-parser/internal/domain"
	"os"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer() KafkaProducer {
	brokers := os.Getenv("KAFKA_BROKERS")
	topicName := os.Getenv("KAFKA_TOPIC")
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers),
		Topic:                  topicName,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}
	return KafkaProducer{writer: writer}
}

func (p KafkaProducer) sendData(dto domain.ResultDto) {
	b, err := json.Marshal(dto)
	if err != nil {
		slog.Error(err.Error())
	}
	if err := p.writer.WriteMessages(context.Background(), kafka.Message{Key: []byte(dto.Category), Value: b}); err != nil {
		slog.Error(err.Error())
	}
}

func (p KafkaProducer) Close() error {
	slog.Info("closing kafka producer")
	return p.writer.Close()
}
