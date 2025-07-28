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

type GateExchange struct {
	config ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type GateContract struct {
	Name              string `json:"name"`
	FundingRate       string `json:"funding_rate"`
	FundingNextApply  int64  `json:"funding_next_apply"`
	MarkPrice         string `json:"mark_price"`
	IndexPrice        string `json:"index_price"`
	LastPrice         string `json:"last_price"`
	Status            string `json:"status"`
}

func NewGateExchange(config ExchangeConfig, logger *logrus.Logger) *GateExchange {
	return &GateExchange{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *GateExchange) GetName() string {
	return "gate"
}

func (g *GateExchange) IsHealthy() bool {
	url := fmt.Sprintf("%s/api/v4/futures/usdt/contracts", g.config.BaseURL)
	resp, err := g.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (g *GateExchange) GetFundingRates() ([]FundingRate, error) {
	url := fmt.Sprintf("%s/api/v4/futures/usdt/contracts", g.config.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
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

	var contracts []GateContract
	if err := json.Unmarshal(body, &contracts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var rates []FundingRate
	
	for _, contract := range contracts {
		// Skip if contract is not active or funding rate is empty
		if contract.Status != "trading" || contract.FundingRate == "" {
			continue
		}

		fundingRate, err := strconv.ParseFloat(contract.FundingRate, 64)
		if err != nil {
			g.logger.Warnf("Failed to parse funding rate for %s: %v", contract.Name, err)
			continue
		}

		markPrice, _ := strconv.ParseFloat(contract.MarkPrice, 64)
		indexPrice, _ := strconv.ParseFloat(contract.IndexPrice, 64)

		rates = append(rates, FundingRate{
			Symbol:          contract.Name,
			Exchange:        g.GetName(),
			FundingRate:     fundingRate,
			NextFundingTime: time.Unix(contract.FundingNextApply, 0),
			Timestamp:       time.Now(),
			MarkPrice:       markPrice,
			IndexPrice:      indexPrice,
			LastFundingRate: 0, // Not provided in this endpoint
		})
	}

	g.logger.Infof("Retrieved %d funding rates from Gate.io", len(rates))
	return rates, nil
} 