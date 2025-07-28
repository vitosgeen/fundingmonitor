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

type BinanceExchange struct {
	config ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type BinanceFundingRate struct {
	Symbol               string `json:"symbol"`
	MarkPrice           string `json:"markPrice"`
	IndexPrice          string `json:"indexPrice"`
	LastFundingRate     string `json:"lastFundingRate"`
	NextFundingTime     int64  `json:"nextFundingTime"`
	FundingRate         string `json:"fundingRate"`
	FundingRatePrecision string `json:"fundingRatePrecision"`
	Time                int64  `json:"time"`
}

type BinanceFundingRateResponse struct {
	Code string                  `json:"code"`
	Msg  string                  `json:"msg"`
	Data []BinanceFundingRate   `json:"data"`
}

func NewBinanceExchange(config ExchangeConfig, logger *logrus.Logger) *BinanceExchange {
	return &BinanceExchange{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (b *BinanceExchange) GetName() string {
	return "binance"
}

func (b *BinanceExchange) IsHealthy() bool {
	// Simple health check by making a request to the funding rate endpoint
	url := fmt.Sprintf("%s/fapi/v1/premiumIndex", b.config.BaseURL)
	resp, err := b.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (b *BinanceExchange) GetFundingRates() ([]FundingRate, error) {
	url := fmt.Sprintf("%s/fapi/v1/premiumIndex", b.config.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if provided
	if b.config.APIKey != "" {
		req.Header.Set("X-MBX-APIKEY", b.config.APIKey)
	}

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

	var binanceRates []BinanceFundingRate
	if err := json.Unmarshal(body, &binanceRates); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var rates []FundingRate
	for _, rate := range binanceRates {
		fundingRate, err := strconv.ParseFloat(rate.LastFundingRate, 64)
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

		lastFundingRate, err := strconv.ParseFloat(rate.LastFundingRate, 64)
		if err != nil {
			b.logger.Warnf("Failed to parse last funding rate for %s: %v", rate.Symbol, err)
		}

		rates = append(rates, FundingRate{
			Symbol:           rate.Symbol,
			Exchange:         b.GetName(),
			FundingRate:      fundingRate,
			NextFundingTime:  time.Unix(rate.NextFundingTime/1000, 0),
			Timestamp:        time.Now(),
			MarkPrice:        markPrice,
			IndexPrice:       indexPrice,
			LastFundingRate:  lastFundingRate,
		})
	}

	b.logger.Infof("Retrieved %d funding rates from Binance", len(rates))
	return rates, nil
} 