package infrastructure

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"news-parser/internal/domain"
)

const Url = "https://api.gdeltproject.org/api/v2/doc/doc?query=(bitcoin%20OR%20BTC)&mode=artlist&format=json&maxrecords=50&sort=datedesc&timespan=6h"

type NewsRequester struct {
	Client *http.Client
}

func (nr NewsRequester) DoNewsRequest() (domain.NewsArticles, error) {
	req, err := http.NewRequest(http.MethodGet, Url, nil)
	if err != nil {
		return domain.NewsArticles{}, err
	}

	resp, err := nr.Client.Do(req)
	if err != nil {
		return domain.NewsArticles{}, err

	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.NewsArticles{}, err
	}

	articles := domain.NewsArticles{}
	if err := json.Unmarshal(body, &articles); err != nil {
		return domain.NewsArticles{}, err
	}
	return articles, nil
}

func (nr NewsRequester) DoDataRequest(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := nr.Client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
