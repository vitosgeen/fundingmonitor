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

type OKXExchange struct {
	config ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type OKXFundingRate struct {
	InstId        string `json:"instId"`
	InstType      string `json:"instType"`
	FundingRate   string `json:"fundingRate"`
	NextFundingTime string `json:"nextFundingTime"`
	FundingRatePrecision string `json:"fundingRatePrecision"`
	MarkPrice     string `json:"markPx"`
	IndexPrice    string `json:"idxPx"`
	LastFundingRate string `json:"lastFundingRate"`
}

type OKXFundingRateResponse struct {
	Code string               `json:"code"`
	Msg  string               `json:"msg"`
	Data []OKXFundingRate    `json:"data"`
}

func NewOKXExchange(config ExchangeConfig, logger *logrus.Logger) *OKXExchange {
	return &OKXExchange{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (o *OKXExchange) GetName() string {
	return "okx"
}

func (o *OKXExchange) IsHealthy() bool {
	// Simple health check by making a request to the funding rate endpoint
	url := fmt.Sprintf("%s/api/v5/public/funding-rate", o.config.BaseURL)
	resp, err := o.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (o *OKXExchange) GetFundingRates() ([]FundingRate, error) {
	url := fmt.Sprintf("%s/api/v5/public/funding-rate", o.config.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if provided
	if o.config.APIKey != "" {
		req.Header.Set("OK-ACCESS-KEY", o.config.APIKey)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("instType", "SWAP") // Get perpetual swap funding rates
	q.Add("limit", "100")     // Get up to 100 results
	req.URL.RawQuery = q.Encode()

	resp, err := o.client.Do(req)
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

	var okxResponse OKXFundingRateResponse
	if err := json.Unmarshal(body, &okxResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if okxResponse.Code != "0" {
		return nil, fmt.Errorf("OKX API error: %s", okxResponse.Msg)
	}

	var rates []FundingRate
	for _, rate := range okxResponse.Data {
		fundingRate, err := strconv.ParseFloat(rate.FundingRate, 64)
		if err != nil {
			o.logger.Warnf("Failed to parse funding rate for %s: %v", rate.InstId, err)
			continue
		}

		markPrice, err := strconv.ParseFloat(rate.MarkPrice, 64)
		if err != nil {
			o.logger.Warnf("Failed to parse mark price for %s: %v", rate.InstId, err)
		}

		indexPrice, err := strconv.ParseFloat(rate.IndexPrice, 64)
		if err != nil {
			o.logger.Warnf("Failed to parse index price for %s: %v", rate.InstId, err)
		}

		lastFundingRate, err := strconv.ParseFloat(rate.LastFundingRate, 64)
		if err != nil {
			o.logger.Warnf("Failed to parse last funding rate for %s: %v", rate.InstId, err)
		}

		// Parse next funding time
		nextFundingTime, err := strconv.ParseInt(rate.NextFundingTime, 10, 64)
		if err != nil {
			o.logger.Warnf("Failed to parse next funding time for %s: %v", rate.InstId, err)
		}

		rates = append(rates, FundingRate{
			Symbol:           rate.InstId,
			Exchange:         o.GetName(),
			FundingRate:      fundingRate,
			NextFundingTime:  time.Unix(nextFundingTime/1000, 0),
			Timestamp:        time.Now(),
			MarkPrice:        markPrice,
			IndexPrice:       indexPrice,
			LastFundingRate:  lastFundingRate,
		})
	}

	o.logger.Infof("Retrieved %d funding rates from OKX", len(rates))
	return rates, nil
} 