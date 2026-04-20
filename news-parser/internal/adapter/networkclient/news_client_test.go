package networkclient

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"news-parser/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewsRequesterDoNewsRequest(t *testing.T) {
	tests := []struct {
		name          string
		category      domain.Category
		responseCode  int
		responseBody  string
		transportErr  error
		expected      domain.Articles
		expectedError string
	}{
		{
			name:         "success",
			category:     domain.CryptoCategory,
			responseCode: http.StatusOK,
			responseBody: `{"articles":[{"title":"BTC jumps","url":"https://example.com/btc","socialimage":"https://example.com/btc.png"}]}`,
			expected: domain.Articles{Articles: []domain.GdeltAPIDto{{
				Title:       "BTC jumps",
				URL:         "https://example.com/btc",
				SocialImage: "https://example.com/btc.png",
			}}},
		},
		{
			name:          "returns error on transport failure",
			category:      domain.TechnologyCategory,
			transportErr:  errors.New("network down"),
			expectedError: "error doing request to GDELT API",
		},
		{
			name:          "returns error on invalid json",
			category:      domain.EconomyCategory,
			responseCode:  http.StatusOK,
			responseBody:  `{"articles":`,
			expectedError: "error unmarshalling JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedURL := urlByCategory(tt.category)
			require.NotEmpty(t, expectedURL)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, expectedURL, r.Header.Get("X-Original-URL"))
				w.WriteHeader(tt.responseCode)
				_, err := io.WriteString(w, tt.responseBody)
				require.NoError(t, err)
			}))
			defer server.Close()

			client := server.Client()
			client.Transport = rewriteTransport{
				base:        client.Transport,
				targetURL:   server.URL,
				forwardHost: "api.gdeltproject.org",
				returnErr:   tt.transportErr,
			}

			requester := NewsRequester{Client: client}
			result, err := requester.DoNewsRequest(tt.category)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, domain.Articles{}, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewsRequesterDoDataRequest(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		responseCode  int
		responseBody  string
		transportErr  error
		expected      []byte
		expectedError string
	}{
		{
			name:         "success",
			url:          "https://example.com/article",
			responseCode: http.StatusOK,
			responseBody: "<html>bitcoin</html>",
			expected:     []byte("<html>bitcoin</html>"),
		},
		{
			name:          "returns status error",
			url:           "https://example.com/missing",
			responseCode:  http.StatusBadGateway,
			responseBody:  "bad gateway",
			expectedError: "502 Bad Gateway",
		},
		{
			name:          "returns transport error",
			url:           "https://example.com/unreachable",
			transportErr:  errors.New("dial error"),
			expectedError: "error doing data request to url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.url, r.Header.Get("X-Original-URL"))
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

			requester := NewsRequester{Client: client}
			result, err := requester.DoDataRequest(tt.url)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type rewriteTransport struct {
	base        http.RoundTripper
	targetURL   string
	forwardHost string
	returnErr   error
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

	rewritten.URL.Scheme = target.Scheme
	rewritten.URL.Host = target.Host
	rewritten.Host = target.Host
	rewritten.Header.Set("X-Original-URL", req.URL.String())

	if t.forwardHost != "" {
		rewritten.URL.Path = "/"
		rewritten.URL.RawPath = ""
		rewritten.URL.RawQuery = ""
	}

	return t.base.RoundTrip(rewritten)
}

func TestURLByCategory(t *testing.T) {
	tests := []struct {
		name     string
		category domain.Category
		expected string
	}{
		{name: "politics", category: domain.PoliticsCategory, expected: militaryURL},
		{name: "environment", category: domain.EnvironmentCategory, expected: energeticsURL},
		{name: "economy", category: domain.EconomyCategory, expected: economyURL},
		{name: "technology", category: domain.TechnologyCategory, expected: itURL},
		{name: "crypto", category: domain.CryptoCategory, expected: cryptoURL},
		{name: "unknown", category: domain.Category(99), expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, urlByCategory(tt.category))
		})
	}
}

func TestNewsRequesterDoNewsRequestUsesCategoryURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originalURL := r.Header.Get("X-Original-URL")
		assert.True(t, strings.Contains(originalURL, "maxrecords=50") || strings.Contains(originalURL, "maxrecords=100"))
		assert.Contains(t, originalURL, "format=json")
		_, err := io.WriteString(w, `{"articles":[]}`)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := server.Client()
	client.Transport = rewriteTransport{
		base:        client.Transport,
		targetURL:   server.URL,
		forwardHost: "api.gdeltproject.org",
	}

	requester := NewsRequester{Client: client}
	_, err := requester.DoNewsRequest(domain.PoliticsCategory)
	require.NoError(t, err)
}
