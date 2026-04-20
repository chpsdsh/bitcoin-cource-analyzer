package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKafkaConfig(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		expected      KafkaConfig
		expectedError error
	}{
		{
			name: "success",
			env: map[string]string{
				"KAFKA_BROKERS":            "kafka-1:9092,kafka-2:9092",
				"KAFKA_ARTICLES_TOPIC":     "articles",
				"KAFKA_LLM_RESPONSE_TOPIC": "llm-responses",
				"KAFKA_NEWS_TOPIC":         "news",
				"KAFKA_GROUP_ID":           "cache-service",
			},
			expected: KafkaConfig{
				Brokers:               []string{"kafka-1:9092", "kafka-2:9092"},
				InputArticlesTopic:    "articles",
				InputLLMResponseTopic: "llm-responses",
				InputNewsTopic:        "news",
				GroupID:               "cache-service",
			},
		},
		{
			name: "returns error when brokers are missing",
			env: map[string]string{
				"KAFKA_ARTICLES_TOPIC":     "articles",
				"KAFKA_LLM_RESPONSE_TOPIC": "llm-responses",
				"KAFKA_NEWS_TOPIC":         "news",
				"KAFKA_GROUP_ID":           "cache-service",
			},
			expectedError: ErrKafkaBrokersNotSet,
		},
		{
			name: "returns error when articles topic is missing",
			env: map[string]string{
				"KAFKA_BROKERS":            "kafka-1:9092",
				"KAFKA_LLM_RESPONSE_TOPIC": "llm-responses",
				"KAFKA_NEWS_TOPIC":         "news",
				"KAFKA_GROUP_ID":           "cache-service",
			},
			expectedError: ErrKafkaInputArticlesTopicNotSet,
		},
		{
			name: "returns error when llm response topic is missing",
			env: map[string]string{
				"KAFKA_BROKERS":        "kafka-1:9092",
				"KAFKA_ARTICLES_TOPIC": "articles",
				"KAFKA_NEWS_TOPIC":     "news",
				"KAFKA_GROUP_ID":       "cache-service",
			},
			expectedError: ErrKafkaInputLLMResponseTopicNotSet,
		},
		{
			name: "returns error when news topic is missing",
			env: map[string]string{
				"KAFKA_BROKERS":            "kafka-1:9092",
				"KAFKA_ARTICLES_TOPIC":     "articles",
				"KAFKA_LLM_RESPONSE_TOPIC": "llm-responses",
				"KAFKA_GROUP_ID":           "cache-service",
			},
			expectedError: ErrKafkaInputNewsTopicNotSet,
		},
		{
			name: "returns error when group id is missing",
			env: map[string]string{
				"KAFKA_BROKERS":            "kafka-1:9092",
				"KAFKA_ARTICLES_TOPIC":     "articles",
				"KAFKA_LLM_RESPONSE_TOPIC": "llm-responses",
				"KAFKA_NEWS_TOPIC":         "news",
			},
			expectedError: ErrKafkaGroupNotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("KAFKA_BROKERS", "")
			t.Setenv("KAFKA_ARTICLES_TOPIC", "")
			t.Setenv("KAFKA_LLM_RESPONSE_TOPIC", "")
			t.Setenv("KAFKA_NEWS_TOPIC", "")
			t.Setenv("KAFKA_GROUP_ID", "")

			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			cfg, err := NewKafkaConfig()

			if tt.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedError)
				assert.Equal(t, KafkaConfig{}, cfg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
