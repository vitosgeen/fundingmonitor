package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type BybitExchange struct {
	config ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type BybitTicker struct {
	Symbol           string `json:"symbol"`
	FundingRate     string `json:"fundingRate"`
	MarkPrice       string `json:"markPrice"`
	IndexPrice      string `json:"indexPrice"`
	NextFundingTime string `json:"nextFundingTime"`
}

type BybitTickerResponse struct {
	RetCode int                    `json:"retCode"`
	RetMsg  string                 `json:"retMsg"`
	Result  struct {
		List []BybitTicker `json:"list"`
	} `json:"result"`
}

func NewBybitExchange(config ExchangeConfig, logger *logrus.Logger) *BybitExchange {
	return &BybitExchange{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (b *BybitExchange) GetName() string {
	return "bybit"
}

func (b *BybitExchange) IsHealthy() bool {
	// Simple health check by making a request to the tickers endpoint
	url := fmt.Sprintf("%s/v5/market/tickers", b.config.BaseURL)
	resp, err := b.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (b *BybitExchange) GetFundingRates() ([]FundingRate, error) {
	url := fmt.Sprintf("%s/v5/market/tickers", b.config.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if provided
	if b.config.APIKey != "" {
		req.Header.Set("X-BAPI-API-KEY", b.config.APIKey)
	}

	// Add query parameters for current funding rates
	q := req.URL.Query()
	q.Add("category", "linear")
	req.URL.RawQuery = q.Encode()

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var bybitResponse BybitTickerResponse
	if err := json.Unmarshal(body, &bybitResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if bybitResponse.RetCode != 0 {
		return nil, fmt.Errorf("Bybit API error: %s", bybitResponse.RetMsg)
	}

	var rates []FundingRate
	for _, rate := range bybitResponse.Result.List {
		fundingRate, err := strconv.ParseFloat(rate.FundingRate, 64)
		if err != nil {
			b.logger.Warnf("Failed to parse funding rate for %s: %v", rate.Symbol, err)
			continue
		}

		markPrice, err := strconv.ParseFloat(rate.MarkPrice, 64)
		if err != nil {
			b.logger.Warnf("Failed to parse mark price for %s: %v", rate.Symbol, err)
		}

		indexPrice, err := strconv.ParseFloat(rate.IndexPrice, 64)
		if err != nil {
			b.logger.Warnf("Failed to parse index price for %s: %v", rate.Symbol, err)
		}

		// Parse next funding time
		nextFundingTime, err := strconv.ParseInt(rate.NextFundingTime, 10, 64)
		if err != nil {
			b.logger.Warnf("Failed to parse next funding time for %s: %v", rate.Symbol, err)
		}

		rates = append(rates, FundingRate{
			Symbol:           rate.Symbol,
			Exchange:         b.GetName(),
			FundingRate:      fundingRate,
			NextFundingTime:  time.Unix(nextFundingTime/1000, 0),
			Timestamp:        time.Now(),
			MarkPrice:        markPrice,
			IndexPrice:       indexPrice,
		})
	}

	b.logger.Infof("Retrieved %d funding rates from Bybit", len(rates))
	return rates, nil
} 