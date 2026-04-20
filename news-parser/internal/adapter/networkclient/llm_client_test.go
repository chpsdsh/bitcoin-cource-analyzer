package networkclient

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMClientStartLLMPrediction(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		responseCode  int
		transportErr  error
		expectedError string
	}{
		{
			name:         "success",
			responseCode: http.StatusOK,
		},
		{
			name:          "returns error on non ok status",
			responseCode:  http.StatusBadGateway,
			expectedError: "status code is not 200 status: 502",
		},
		{
			name:          "returns error on transport failure",
			transportErr:  errors.New("connection refused"),
			expectedError: "error doing request to LLM",
		},
		{
			name:          "returns error on invalid address",
			address:       "://bad-url",
			expectedError: "error creating request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.address == "" {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodPost, r.Method)
					w.WriteHeader(tt.responseCode)
				}))
				defer server.Close()
				tt.address = server.URL
			}

			client := http.DefaultClient
			if tt.transportErr != nil {
				client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					return nil, tt.transportErr
				})}
			}

			llmClient := LLMClient{
				Client:     client,
				LLMAddress: tt.address,
			}

			err := llmClient.StartLLMPrediction()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
