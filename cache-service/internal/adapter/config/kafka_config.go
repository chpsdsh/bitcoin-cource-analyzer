package config

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrKafkaBrokersNotSet               = errors.New("kafka brokers should be set with KAFKA_BROKERS env variable")
	ErrKafkaInputNewsTopicNotSet        = errors.New("kafka input topic should be set with KAFKA_NEWS_TOPIC env variable")
	ErrKafkaGroupNotSet                 = errors.New("kafka group should be set with KAFKA_GROUP_ID env variable")
	ErrKafkaInputArticlesTopicNotSet    = errors.New("kafka input topic should be set with KAFKA_ARTICLES_TOPIC env variable")
	ErrKafkaInputLLMResponseTopicNotSet = errors.New("kafka input topic should be set with KAFKA_LLM_RESPONSE_TOPIC env variable")
)

type KafkaConfig struct {
	Brokers               []string
	InputNewsTopic        string
	InputArticlesTopic    string
	InputLLMResponseTopic string
	GroupID               string
}

func NewKafkaConfig() (KafkaConfig, error) {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		return KafkaConfig{}, ErrKafkaBrokersNotSet
	}

	inputArticlesTopic := os.Getenv("KAFKA_ARTICLES_TOPIC")
	if inputArticlesTopic == "" {
		return KafkaConfig{}, ErrKafkaInputArticlesTopicNotSet
	}

	inputLLMResponseTopic := os.Getenv("KAFKA_LLM_RESPONSE_TOPIC")
	if inputLLMResponseTopic == "" {
		return KafkaConfig{}, ErrKafkaInputLLMResponseTopicNotSet
	}

	inputNewsTopic := os.Getenv("KAFKA_NEWS_TOPIC")
	if inputNewsTopic == "" {
		return KafkaConfig{}, ErrKafkaInputNewsTopicNotSet
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		return KafkaConfig{}, ErrKafkaGroupNotSet
	}

	brokers := strings.Split(brokersEnv, ",")

	return KafkaConfig{
		Brokers:               brokers,
		InputNewsTopic:        inputNewsTopic,
		InputArticlesTopic:    inputArticlesTopic,
		InputLLMResponseTopic: inputLLMResponseTopic,
		GroupID:               groupID,
	}, nil
}
