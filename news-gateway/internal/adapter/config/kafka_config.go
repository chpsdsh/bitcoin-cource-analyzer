package config

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrKafkaBrokersNotSet = errors.New("kafka brokers should be set with KAFKA_BROKERS env variable")
	ErrKafkaTopicNotSet   = errors.New("kafka topic should be set with KAFKA_NEWS_TOPIC env variable")
	ErrKafkaGroupNotSet   = errors.New("kafka group should be set with KAFKA_GROUP_ID env variable")
)

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewKafkaConfig() (KafkaConfig, error) {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		return KafkaConfig{}, ErrKafkaBrokersNotSet
	}

	topic := os.Getenv("KAFKA_NEWS_TOPIC")
	if topic == "" {
		return KafkaConfig{}, ErrKafkaTopicNotSet
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		return KafkaConfig{}, ErrKafkaGroupNotSet
	}

	brokers := strings.Split(brokersEnv, ",")

	return KafkaConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	}, nil
}
