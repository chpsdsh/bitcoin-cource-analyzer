package networkclient

import (
	"context"
	"fmt"
	"net/http"
)

type LLMClient struct {
	Client     *http.Client
	LLMAddress string
}

func (c LLMClient) StartLLMPrediction() error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.LLMAddress, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error doing request to LLM: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code is not 200 status: %d", resp.StatusCode)
	}

	return nil
}
