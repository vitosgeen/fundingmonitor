package infrastructure

import (
	"encoding/json"
	"fmt"
	"fundingmonitor/internal/domain"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type MEXCClient struct {
	config domain.ExchangeConfig
	logger *logrus.Logger
	client *http.Client
}

type MEXCFundingRate struct {
	Symbol         string  `json:"symbol"`
	FundingRate    float64 `json:"fundingRate"`
	MaxFundingRate float64 `json:"maxFundingRate"`
	MinFundingRate float64 `json:"minFundingRate"`
	CollectCycle   int     `json:"collectCycle"`
	NextSettleTime int64   `json:"nextSettleTime"`
	Timestamp      int64   `json:"timestamp"`
	MarkPrice      float64 `json:"markPrice"`
	IndexPrice     float64 `json:"indexPrice"`
}

type MEXCFundingRateResponse struct {
	Success bool              `json:"success"`
	Code    int               `json:"code"`
	Msg     string            `json:"msg"`
	Data    []MEXCFundingRate `json:"data"`
}

func NewMEXCClient(config domain.ExchangeConfig, logger *logrus.Logger) *MEXCClient {
	return &MEXCClient{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (m *MEXCClient) GetName() string {
	return "mexc"
}

func (m *MEXCClient) IsHealthy() bool {
	url := fmt.Sprintf("%s/api/v1/contract/funding_rate", m.config.BaseURL)
	resp, err := m.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (m *MEXCClient) GetFundingRates() ([]domain.FundingRate, error) {
	url := fmt.Sprintf("%s/api/v1/contract/funding_rate", m.config.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.client.Do(req)
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

	var mexcResponse MEXCFundingRateResponse
	if err := json.Unmarshal(body, &mexcResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !mexcResponse.Success && mexcResponse.Code != 0 {
		return nil, fmt.Errorf("MEXC API error: %s", mexcResponse.Msg)
	}

	var rates []domain.FundingRate
	for _, rate := range mexcResponse.Data {
		rates = append(rates, domain.FundingRate{
			Symbol:          rate.Symbol,
			Exchange:        m.GetName(),
			FundingRate:     rate.FundingRate,
			NextFundingTime: time.Unix(rate.NextSettleTime/1000, 0),
			Timestamp:       time.Unix(rate.Timestamp/1000, 0),
			MarkPrice:       0, // MEXC doesn't provide mark price in this endpoint
			IndexPrice:      0, // MEXC doesn't provide index price in this endpoint
			LastFundingRate: 0, // MEXC doesn't provide last funding rate in this endpoint
		})
	}

	m.logger.Infof("Retrieved %d funding rates from MEXC", len(rates))
	return rates, nil
}
