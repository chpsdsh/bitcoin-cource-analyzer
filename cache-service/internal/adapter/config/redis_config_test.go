package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisConfig(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		expected      RedisConfig
		expectedError error
	}{
		{
			name: "success",
			env: map[string]string{
				"REDIS_ADDR":            "localhost:6379",
				"REDIS_PASSWORD":        "secret",
				"REDIS_ARTICLES_DB":     "1",
				"REDIS_NEWS_DB":         "2",
				"REDIS_LLM_RESPONSE_DB": "3",
			},
			expected: RedisConfig{
				RedisAddr:          "localhost:6379",
				RedisPassword:      "secret",
				RedisArticlesDB:    1,
				RedisNewsDB:        2,
				RedisLLMResponseDB: 3,
			},
		},
		{
			name: "returns error when redis addr is missing",
			env: map[string]string{
				"REDIS_PASSWORD":        "secret",
				"REDIS_ARTICLES_DB":     "1",
				"REDIS_NEWS_DB":         "2",
				"REDIS_LLM_RESPONSE_DB": "3",
			},
			expectedError: ErrRedisAddrNotSet,
		},
		{
			name: "returns error when redis password is missing",
			env: map[string]string{
				"REDIS_ADDR":            "localhost:6379",
				"REDIS_ARTICLES_DB":     "1",
				"REDIS_NEWS_DB":         "2",
				"REDIS_LLM_RESPONSE_DB": "3",
			},
			expectedError: ErrRedisPasswordNotSet,
		},
		{
			name: "returns joined error when articles db is invalid",
			env: map[string]string{
				"REDIS_ADDR":            "localhost:6379",
				"REDIS_PASSWORD":        "secret",
				"REDIS_ARTICLES_DB":     "bad",
				"REDIS_NEWS_DB":         "2",
				"REDIS_LLM_RESPONSE_DB": "3",
			},
			expectedError: ErrRedisNewsDBNotSet,
		},
		{
			name: "returns joined error when news db is invalid",
			env: map[string]string{
				"REDIS_ADDR":            "localhost:6379",
				"REDIS_PASSWORD":        "secret",
				"REDIS_ARTICLES_DB":     "1",
				"REDIS_NEWS_DB":         "bad",
				"REDIS_LLM_RESPONSE_DB": "3",
			},
			expectedError: ErrRedisNewsDBNotSet,
		},
		{
			name: "returns joined error when llm response db is invalid",
			env: map[string]string{
				"REDIS_ADDR":            "localhost:6379",
				"REDIS_PASSWORD":        "secret",
				"REDIS_ARTICLES_DB":     "1",
				"REDIS_NEWS_DB":         "2",
				"REDIS_LLM_RESPONSE_DB": "bad",
			},
			expectedError: ErrRedisLLMResponseDBNotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_ADDR", "")
			t.Setenv("REDIS_PASSWORD", "")
			t.Setenv("REDIS_ARTICLES_DB", "")
			t.Setenv("REDIS_NEWS_DB", "")
			t.Setenv("REDIS_LLM_RESPONSE_DB", "")

			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			cfg, err := NewRedisConfig()

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Equal(t, RedisConfig{}, cfg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
