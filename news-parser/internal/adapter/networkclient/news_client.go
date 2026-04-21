package networkclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"news-parser/internal/domain"
	"news-parser/internal/observability"
)

const (
	militaryURL   = "https://api.gdeltproject.org/api/v2/doc/doc?query=(sanctions%20OR%20war%20OR%20election%20OR%20military%20OR%20ceasefire)%20(theme:WB_2433_CONFLICT_AND_VIOLENCE%20OR%20theme:ARMEDCONFLICT%20OR%20theme:LEGISLATION%20OR%20theme:ELECTION%20OR%20theme:TERROR)%20sourcelang:english&mode=artlist&format=json&maxrecords=50&sort=datedesc&timespan=6h"
	energeticsURL = "https://api.gdeltproject.org/api/v2/doc/doc?query=(oil%20OR%20%22natural%20gas%22%20OR%20%22electricity%20prices%22%20OR%20OPEC%20OR%20silicon)%20(theme:WB_507_ENERGY_AND_EXTRACTIVES%20OR%20theme:WB_895_MINING_SYSTEMS%20OR%20theme:ENV_OIL%20OR%20theme:WB_566_ENVIRONMENT_AND_NATURAL_RESOURCES%20OR%20theme:DISASTER_FIRE)%20sourcelang:english&mode=artlist&format=json&maxrecords=50&sort=datedesc&timespan=6h"
	economyURL    = "https://api.gdeltproject.org/api/v2/doc/doc?query=(inflation%20OR%20%22interest%20rates%22%20OR%20Fed%20OR%20recession%20OR%20%22dollar%20index%22)%20(theme:EPU_POLICY%20OR%20theme:EPU_ECONOMY%20OR%20theme:ECON_STOCKMARKET%20OR%20theme:ECON_TAXATION%20OR%20theme:WB_450_DEBT)%20sourcelang:english&mode=artlist&format=json&maxrecords=50&sort=datedesc&timespan=6h"
	itURL         = "https://api.gdeltproject.org/api/v2/doc/doc?query=(cyberattack%20OR%20exploit%20OR%20ransomware%20OR%20outage%20OR%20GPU)%20(theme:WB_667_ICT_INFRASTRUCTURE%20OR%20theme:WB_652_ICT_APPLICATIONS%20OR%20theme:WB_670_ICT_SECURITY%20OR%20theme:WB_669_SOFTWARE_INFRASTRUCTURE%20OR%20theme:CYBER_ATTACK)%20sourcelang:english&mode=artlist&format=json&maxrecords=50&sort=datedesc&timespan=6h"
	cryptoURL     = "https://api.gdeltproject.org/api/v2/doc/doc?query=(BTC%20%20OR%20blockchain%20OR%20Bitcoin%20OR%20ETF)%20(theme:EPU_CATS_REGULATION%20OR%20theme:WB_328_FINANCIAL_INTEGRITY%20OR%20theme:WB_332_CAPITAL_MARKETS%20OR%20theme:WB_1234_BANKING_INSTITUTIONS%20OR%20theme:WB_336_NON_BANK_FINANCIAL_INSTITUTIONS)%20sourcelang:english&mode=artlist&format=json&maxrecords=100&sort=datedesc&timespan=6h"
)

type NewsRequester struct {
	Client *http.Client
}

func (nr NewsRequester) DoNewsRequest(ctx context.Context, category domain.Category) (domain.Articles, error) {
	URL := urlByCategory(category)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return domain.Articles{}, fmt.Errorf("error creating request %w", err)
	}
	if traceID := observability.TraceIDFromContext(ctx); traceID != "" {
		req.Header.Set(observability.TraceIDHeader, traceID)
	}

	resp, err := nr.Client.Do(req)
	if err != nil {
		return domain.Articles{}, fmt.Errorf("error doing request to GDELT API %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Articles{}, fmt.Errorf("error reading request body %w", err)
	}

	articles := domain.Articles{}
	if err = json.Unmarshal(body, &articles); err != nil {
		return domain.Articles{}, fmt.Errorf("error unmarshalling JSON %w", err)
	}
	return articles, nil
}

func (nr NewsRequester) DoDataRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request %w", err)
	}
	if traceID := observability.TraceIDFromContext(ctx); traceID != "" {
		req.Header.Set(observability.TraceIDHeader, traceID)
	}

	resp, err := nr.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing data request to url %s %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON %w", err)
	}
	return body, nil
}

func urlByCategory(category domain.Category) string {
	var URL string
	switch category {
	case domain.PoliticsCategory:
		URL = militaryURL
	case domain.EnvironmentCategory:
		URL = energeticsURL
	case domain.CryptoCategory:
		URL = cryptoURL
	case domain.EconomyCategory:
		URL = economyURL
	case domain.TechnologyCategory:
		URL = itURL
	}
	return URL
}
