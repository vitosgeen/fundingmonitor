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

type XTExchange struct {
	config ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type XTFundingRate struct {
	Symbol        string `json:"symbol"`
	FundingRate   string `json:"fundingRate"`
	NextFundingTime int64 `json:"nextFundingTime"`
	Timestamp     int64  `json:"timestamp"`
}

type XTFundingRateResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    []XTFundingRate `json:"data"`
}

func NewXTExchange(config ExchangeConfig, logger *logrus.Logger) *XTExchange {
	return &XTExchange{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (x *XTExchange) GetName() string {
	return "xt"
}

func (x *XTExchange) IsHealthy() bool {
	url := fmt.Sprintf("%s/v4/public/funding-rate", x.config.BaseURL)
	resp, err := x.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (x *XTExchange) GetFundingRates() ([]FundingRate, error) {
	url := fmt.Sprintf("%s/v4/public/funding-rate", x.config.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := x.client.Do(req)
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

	var xtResponse XTFundingRateResponse
	if err := json.Unmarshal(body, &xtResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if xtResponse.Code != 0 {
		return nil, fmt.Errorf("XT API error: %s", xtResponse.Message)
	}

	var rates []FundingRate
	for _, rate := range xtResponse.Data {
		fundingRate, err := strconv.ParseFloat(rate.FundingRate, 64)
		if err != nil {
			x.logger.Warnf("Failed to parse funding rate for %s: %v", rate.Symbol, err)
			continue
		}

		rates = append(rates, FundingRate{
			Symbol:          rate.Symbol,
			Exchange:        x.GetName(),
			FundingRate:     fundingRate,
			NextFundingTime: time.Unix(rate.NextFundingTime/1000, 0),
			Timestamp:       time.Unix(rate.Timestamp/1000, 0),
			MarkPrice:       0,
			IndexPrice:      0,
			LastFundingRate: 0,
		})
	}

	x.logger.Infof("Retrieved %d funding rates from XT.com", len(rates))
	return rates, nil
} 