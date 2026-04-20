package binanceclient

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"llm-consumer/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinanceClientRequestBTCPrice(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  string
		transportErr  error
		expected      domain.PriceResponse
		expectedError string
	}{
		{
			name:         "success",
			responseCode: http.StatusOK,
			responseBody: `{"price":"105000.12"}`,
			expected:     domain.PriceResponse{Price: "105000.12"},
		},
		{
			name:          "returns error on transport failure",
			transportErr:  errors.New("connection refused"),
			expectedError: "could not send request",
		},
		{
			name:          "returns error on invalid json",
			responseCode:  http.StatusOK,
			responseBody:  `{"price":`,
			expectedError: "could not unmarshal response body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, btcURL, r.Header.Get("X-Original-URL"))
				w.WriteHeader(tt.responseCode)
				_, err := io.WriteString(w, tt.responseBody)
				require.NoError(t, err)
			}))
			defer server.Close()

			client := server.Client()
			client.Transport = rewriteTransport{
				base:      client.Transport,
				targetURL: server.URL,
				returnErr: tt.transportErr,
			}

			binanceClient := BinanceClient{httpClient: client}
			result, err := binanceClient.RequestBTCPrice()

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, domain.PriceResponse{}, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewBinanceClient(t *testing.T) {
	client := NewBinanceClient()
	require.NotNil(t, client)
	require.NotNil(t, client.httpClient)
	assert.Equal(t, clientTimeout, client.httpClient.Timeout)
}

type rewriteTransport struct {
	base      http.RoundTripper
	targetURL string
	returnErr error
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.returnErr != nil {
		return nil, t.returnErr
	}

	rewritten := req.Clone(req.Context())
	target, err := url.Parse(t.targetURL)
	if err != nil {
		return nil, err
	}

	rewritten.Header.Set("X-Original-URL", req.URL.String())
	rewritten.URL.Scheme = target.Scheme
	rewritten.URL.Host = target.Host
	rewritten.Host = target.Host

	return t.base.RoundTrip(rewritten)
}
