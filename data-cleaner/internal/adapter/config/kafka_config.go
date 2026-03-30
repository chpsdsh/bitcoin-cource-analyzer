package config

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrKafkaBrokersNotSet     = errors.New("kafka brokers should be set with KAFKA_BROKERS env variable")
	ErrKafkaInputTopicNotSet  = errors.New("kafka input topic should be set with KAFKA_ARTICLES_TOPIC env variable")
	ErrKafkaGroupNotSet       = errors.New("kafka group should be set with KAFKA_GROUP_ID env variable")
	ErrKafkaOutputTopicNotSet = errors.New("kafka output topic should be set with KAFKA_ARTICLES_TOPIC_CLEAN env variable")
)

type KafkaConfig struct {
	Brokers     []string
	InputTopic  string
	OutputTopic string
	GroupID     string
}

func NewKafkaConfig() (KafkaConfig, error) {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		return KafkaConfig{}, ErrKafkaBrokersNotSet
	}

	inputTopic := os.Getenv("KAFKA_ARTICLES_TOPIC")
	if inputTopic == "" {
		return KafkaConfig{}, ErrKafkaInputTopicNotSet
	}

	outputTopic := os.Getenv("KAFKA_ARTICLES_TOPIC_CLEAN")
	if outputTopic == "" {
		return KafkaConfig{}, ErrKafkaOutputTopicNotSet
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		return KafkaConfig{}, ErrKafkaGroupNotSet
	}

	brokers := strings.Split(brokersEnv, ",")

	return KafkaConfig{
		Brokers:     brokers,
		InputTopic:  inputTopic,
		OutputTopic: outputTopic,
		GroupID:     groupID,
	}, nil
}
