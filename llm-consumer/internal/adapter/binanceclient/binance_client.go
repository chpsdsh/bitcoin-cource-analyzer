package binanceclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"llm-consumer/internal/domain"
)

const (
	clientTimeout = 10 * time.Second
	btcURL        = "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT"
)

type BinanceClient struct {
	httpClient *http.Client
	priceURL   string
}

func NewBinanceClient() *BinanceClient {
	client := &http.Client{Timeout: clientTimeout}
	return &BinanceClient{
		httpClient: client,
		priceURL:   priceURLFromEnv(),
	}
}

func (c BinanceClient) RequestBTCPrice() (domain.PriceResponse, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.requestURL(), nil)
	if err != nil {
		return domain.PriceResponse{}, fmt.Errorf("could not create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.PriceResponse{}, fmt.Errorf("could not send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.PriceResponse{}, fmt.Errorf("could not read response body: %w", err)
	}

	priceResponse := domain.PriceResponse{}
	if err = json.Unmarshal(data, &priceResponse); err != nil {
		return domain.PriceResponse{}, fmt.Errorf("could not unmarshal response body: %w", err)
	}
	return priceResponse, nil
}

func (c BinanceClient) requestURL() string {
	if c.priceURL == "" {
		return btcURL
	}
	return c.priceURL
}

func priceURLFromEnv() string {
	priceURL := os.Getenv("BINANCE_BTC_PRICE_URL")
	if priceURL == "" {
		return btcURL
	}
	return priceURL
}
