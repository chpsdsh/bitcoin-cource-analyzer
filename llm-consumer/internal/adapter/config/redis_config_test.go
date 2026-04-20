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
				"REDIS_LLM_RESPONSE_DB": "4",
			},
			expected: RedisConfig{
				RedisAddr:     "localhost:6379",
				RedisPassword: "secret",
				RedisDB:       4,
			},
		},
		{
			name: "returns error when redis addr is missing",
			env: map[string]string{
				"REDIS_PASSWORD":        "secret",
				"REDIS_LLM_RESPONSE_DB": "4",
			},
			expectedError: ErrRedisAddrNotSet,
		},
		{
			name: "returns error when redis password is missing",
			env: map[string]string{
				"REDIS_ADDR":            "localhost:6379",
				"REDIS_LLM_RESPONSE_DB": "4",
			},
			expectedError: ErrRedisPasswordNotSet,
		},
		{
			name: "returns joined error when redis db is missing",
			env: map[string]string{
				"REDIS_ADDR":     "localhost:6379",
				"REDIS_PASSWORD": "secret",
			},
			expectedError: ErrRedisDBNotSet,
		},
		{
			name: "returns joined error when redis db is invalid",
			env: map[string]string{
				"REDIS_ADDR":            "localhost:6379",
				"REDIS_PASSWORD":        "secret",
				"REDIS_LLM_RESPONSE_DB": "not-a-number",
			},
			expectedError: ErrRedisDBNotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REDIS_ADDR", "")
			t.Setenv("REDIS_PASSWORD", "")
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
